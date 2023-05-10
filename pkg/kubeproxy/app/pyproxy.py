import subprocess
import sys, os
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(BASE_DIR, '../helper'))
import utils, const
import logging
import random
import time
import prettytable

logging.basicConfig(format='%(asctime)s - %(message)s', level=logging.INFO)

service_clusterIP_prefix = const.service_clusterIP_prefix
default_iptables_path = const.service_iptables_save_path


def alloc_service_clusterIP(service_dict: dict):

    max_alloc_num = 1000  # if exceed this num, that might be not enough service ip
    ip_allocated = set()
    ip = ''
    for service_name in service_dict['services_list']:
        ip = service_dict[service_name].get('clusterIP')
        if ip is not None and ip != '':
            ip_allocated.add(ip)

    while max_alloc_num > 0:
        max_alloc_num -= 1
        # service ip should be like '192.168.xx.xx'
        num0 = service_clusterIP_prefix[0] if len(service_clusterIP_prefix) >= 1 else random.randint(0, 255)
        num1 = service_clusterIP_prefix[1] if len(service_clusterIP_prefix) >= 2 else random.randint(0, 255)
        num2 = service_clusterIP_prefix[2] if len(service_clusterIP_prefix) >= 3 else random.randint(0, 255)
        num3 = service_clusterIP_prefix[3] if len(service_clusterIP_prefix) >= 4 else random.randint(0, 255)
        ip = '.'.join([str(num0), str(num1), str(num2), str(num3)])
        if ip not in ip_allocated:
            break
    if max_alloc_num <= 0:
        logging.error('No available service cluster ip address')
        return ip, False
    return ip, True


def init_iptables():
    """
    init iptables for minik8s, create some necessary chains and insert some necessary rules
    reference to: https://www.bookstack.cn/read/source-code-reading-notes/kubernetes-kube_proxy_iptables.md
    :return: None
    """
    # utils.exec_command(command="iptables-save < ./sources/iptables", shell=True)

    """ In table `nat`, set policy for some chains """
    iptables = dict()
    iptables['chains'] = list()
    iptables['rules'] = list()

    utils.policy_chain('nat', 'PREROUTING', ['ACCEPT'])
    utils.policy_chain('nat', 'INPUT', ['ACCEPT'])
    utils.policy_chain('nat', 'OUTPUT', ['ACCEPT'])
    utils.policy_chain('nat', 'POSTROUTING', ['ACCEPT'])

    """ In table `nat`, create some new chains """
    iptables['chains'].append(utils.create_chain('nat', 'KUBE-SERVICES'))
    # iptables['chains'].append(utils.create_chain('nat', 'KUBE-NODEPORTS'))
    iptables['chains'].append(utils.create_chain('nat', 'KUBE-POSTROUTING'))
    iptables['chains'].append(utils.create_chain('nat', 'KUBE-MARK-MASQ'))
    # iptables['chains'].append(utils.create_chain('nat', 'KUBE-MARK-DROP'))

    """ In table `nat`, add some rule into chains """
    iptables['rules'].append(
        utils.append_rule('nat', 'PREROUTING',
                          utils.make_rulespec(
                              jump='KUBE-SERVICES',
                              comment='kubernetes service portals'
                          ),
                          utils.make_target_extensions())
    )
    iptables['rules'].append(
        utils.append_rule('nat', 'OUTPUT',
                          utils.make_rulespec(
                              jump='KUBE-SERVICES',
                              comment='kubernetes service portals'
                          ),
                          utils.make_target_extensions())
    )
    iptables['rules'].append(
        utils.append_rule('nat', 'POSTROUTING',
                          utils.make_rulespec(
                              jump='KUBE-POSTROUTING',
                              comment='kubernetes postrouting rules'
                          ),
                          utils.make_target_extensions())
    )

    iptables['rules'].append(
        utils.insert_rule('nat', 'KUBE-MARK-MASQ', 1,
                          utils.make_rulespec(
                              jump='MARK'
                          ),
                          utils.make_target_extensions(
                              ormark='0x4000'
                          ))
    )
    iptables['rules'].append(
        utils.insert_rule('nat', 'KUBE-POSTROUTING', 1,
                          utils.make_rulespec(
                              comment='kubernetes service traffic requiring SNAT'
                          ),
                          utils.make_target_extensions(
                              mark='0x4000/0x4000'
                          )
                          )
    )



  


def create_service(service_config: dict, pods_dict: dict, simulate=False):
    """
    used for create a new service using original config file
    :param service_config: dict {'kind': str, 'name': str, 'type': str,
        'selector': dict, 'ports': list, 'instance_name': str,
        'pod_instances': list, 'clusterIP': str}
    :param pods_dict: dict {'chain': list, 'rule': list}
    :return: None
    """
    iptables = dict()
    iptables['chains'] = list()
    iptables['rules'] = list()

    cluster_ip = service_config['clusterIP']
    service_name = service_config['name']
    ports = service_config['ports']
    pod_ip_list = list()
    for pod_instance in service_config['pod_instances']:
        pod_ip_list.append(pods_dict[pod_instance]['ip'])
    strategy = service_config['strategy'] if service_config.get('strategy') is not None else 'random'  # 'random' or 'roundrobin'

    for eachports in ports:
        port = eachports['port']
        targetPort = eachports['targetPort']
        protocol = eachports['protocol']
        set_iptables_clusterIP(cluster_ip=cluster_ip, service_name=service_name,
                               port=port, target_port=targetPort, protocol=protocol,
                               pod_ip_list=pod_ip_list, strategy=strategy, iptables=iptables,
                               simulate=simulate)
    service_config['iptables'] = iptables
    service_config['status'] = 'Running'
    logging.info('Service [%s] ClusterIP [%s] Running Successfully!'
                 % (service_name, cluster_ip))


def rm_service(service_config: dict, simulate=False):
    """
    delete original iptables chains and rules
    :param service_config: dict {'kind': str, 'name': str, 'type': str,
        'selector': dict, 'ports': list, 'instance_name': str,
        'pod_instances': list, 'clusterIP': str}
    :return: None
    """
    # delete original chains and rules
    iptables = service_config['iptables']
    rules = iptables['rules']
    for rule in rules:
        utils.delete_rule_by_spec(table=rule['table'],
                                  chain=rule['chain'],
                                  rulespec=rule['rule-specification'],
                                  simulate=simulate)
    chains = iptables['chains']
    for chain in chains:
        utils.delete_chain(chain['table'], chain['chain'], simulate=simulate)
    service_config['status'] = 'Removed'
    return True


def sync_service(service_config: dict, simulate=False):
    """
    synchronous service iptables with remote master etcd
    :param service_config: master etcd
    :param simulate:
    :return:
    """
    iptables = service_config['iptables']
    chains = iptables['chains']
    for chain in chains:
        utils.create_chain(table=chain['table'], chain=chain['chain'])
    rules = iptables['rules']
    for rule in rules:
        utils.insert_rule(table=rule['table'], chain=rule['chain'], rulenum=1,
                          rulespec=rule['rule-specification'], target_extension=[])


def restart_service(service_config: dict, pods_dict: dict, simulate=False):
    """
    used for restart an exist service, simply
    delete all of the original iptable chains and rules
    then use create_service..
    :param service_config: dict {'kind': str, 'name': str, 'type': str,
        'selector': dict, 'ports': list, 'instance_name': str,
        'pod_instances': list, 'clusterIP': str}
    :param pods_dict: dict {'chain': list, 'rule': list}
    :return: a flag indicating whether change the service
    """
    # compare iplist hash with current ip list
    pod_ip_list = list()
    for pod_instance in service_config['pod_instances']:
        pod_ip_list.append(pods_dict[pod_instance]['ip'])
    # delete original chains and rules
    rm_service(service_config, simulate=simulate)
    # restart this service using create_service
    create_service(service_config, pods_dict, simulate=simulate)
    return


def describe_service(service_config: dict, service_instance_name: str, tb=None, show=False):
    """
    describe a service showing its info
    | name | status | created time | type | cluster IP | external IP | port(s) |
    :param service_config: service config from etcd
    :param service_instance_name: service instance name with its suffix
    :param tb: pretty table used for print beautifully
    :param show: a flag indicating whether to show the bar
    :return: None
    """
    if tb is None:
        tb = prettytable.PrettyTable()
        tb.field_names = ['name', 'instance_name', 'status', 'created time',
                          'type', 'cluster IP', "external IP",
                          'port(s)', 'pod_instances']
    created_time = int(time.time() - service_config['created_time'])
    created_time = str(created_time // 60) + "m" + str(created_time % 60) + 's'
    name = service_config['name'] if service_config.get('name') is not None else '-'
    service_status = service_config['status'] if service_config.get('status') is not None else '-'
    type = '<none>' if service_config.get('type') is None else service_config['type']
    clusterIP = '<none>' if service_config.get('clusterIP') is None else service_config['clusterIP']
    externalIP = '<none>' if service_config.get('externalIP') is None else service_config['externalIP']
    ports: list = service_config.get('ports')
    show_ports = list()
    pod_instances = service_config['pod_instances'] if service_config.get('pod_instances') is not None else list()
    if ports is not None:
        for p in ports:
            format = '%d->%d/%s' % (p['port'], p['targetPort'], p['protocol'])
            show_ports.append(format)
    show_ports = ','.join(show_ports)
    tb.add_row([name, service_instance_name, service_status, created_time.strip(),
                type, clusterIP, externalIP, show_ports, pod_instances])
    if show is True:
        print(tb)


def show_services(service_dict: dict):
    """
    get all services running state
    :param service_dict:
    :return: a list of service running state
    """
    tb = prettytable.PrettyTable()
    tb.field_names = ['name', 'instance_name', 'status', 'created time',
                      'type', 'cluster IP', "external IP",
                      'port(s)', 'pod_instances']

    for service_instance_name in service_dict['services_list']:
        service_config = service_dict[service_instance_name]
        describe_service(service_config=service_config, service_instance_name=service_instance_name, tb=tb, show=False)
    print(tb)


def set_iptables_clusterIP(cluster_ip, service_name, port, target_port, protocol,
                           pod_ip_list, strategy='random', ip_prefix_len=32, iptables: dict = None,
                           simulate=False):
    """
    used for set service clusterIP, only for the first create
    reference to: https://www.bookstack.cn/read/source-code-reading-notes/kubernetes-kube_proxy_iptables.md
    :param cluster_ip: service clusterIP, which should be like xx.xx.xx.xx,
                        don't forget to set security group for that ip address
    :param service_name: service name, only used for comment here
    :param port: exposed service port, which can be visited by other pods by cluster_ip:port
    :param target_port: container runs on target_port actually, must be matched with `pod port`
                        if not matched, we can reject this request or just let it go depending on me
    :param protocol: tcp/udp/all
    :param pod_ip_list: a list of pod ip address, which belongs to the service target pod
    :param strategy: service load balance strategy, which should be random/roundrobin
    :param ip_prefix_len: must be 32 here, so use default value please
    :param iptables: a dict to record each iptable chain and rules create by user
    :return:
    """
    """
    init iptables first, create some necessary chain and rules 
    init_iptables is an idempotent function, which means the effect of
    execute several times equals to the effect of execute one time
    """
    if iptables is None:
        iptables = dict()
        iptables['chains']  = list()
        iptables['rules'] = list()
    kubesvc = 'KUBE-SVC-' + utils.generate_random_str(12, 1)

    iptables['chains'].append(
        utils.create_chain('nat', kubesvc, simulate=simulate)
    )
    iptables['rules'].append(
        utils.insert_rule('nat', 'KUBE-SERVICES', 1,
                          utils.make_rulespec(
                              jump=kubesvc,
                              destination='/'.join([cluster_ip, str(ip_prefix_len)]),
                              protocol=protocol,
                              comment=service_name + ': cluster IP',
                              dport=port
                          ),
                          utils.make_target_extensions(),
                          simulate=simulate)
    )
    iptables['rules'].append(
        utils.insert_rule('nat', 'KUBE-SERVICES', 1,
                          utils.make_rulespec(
                              jump='KUBE-MARK-MASQ',
                              protocol=protocol,
                              destination='/'.join([cluster_ip, str(ip_prefix_len)]),
                              comment=service_name + ': cluster IP',
                              dport=port
                          ),
                          utils.make_target_extensions(),
                          simulate=simulate
                          )
    )

    pod_num = len(pod_ip_list)
    for i in range(pod_num - 1, -1, -1):
        kubesep = 'KUBE-SEP-' + utils.generate_random_str(12, 1)
        iptables['chains'].append(
            utils.create_chain('nat', kubesep, simulate=simulate)
        )

        if strategy == 'random':
            prob = 1 / (pod_num - i)
            if i == pod_num - 1:
                iptables['rules'].append(
                    utils.insert_rule('nat', kubesvc, 1,
                                      utils.make_rulespec(
                                          jump=kubesep
                                      ),
                                      utils.make_target_extensions(),
                                      simulate=simulate)
                )
            else:
                iptables['rules'].append(
                    utils.insert_rule('nat', kubesvc, 1,
                                      utils.make_rulespec(
                                          jump=kubesep,
                                      ),
                                      utils.make_target_extensions(
                                          statistic=True,
                                          mode='random',
                                          probability=prob
                                      ),
                                      simulate=simulate)
                )
        elif strategy == 'roundrobin':
            if i == pod_num - 1:
                iptables['rules'].append(
                    utils.insert_rule('nat', kubesvc, 1,
                                      utils.make_rulespec(
                                          jump=kubesep
                                      ),
                                      utils.make_target_extensions(),
                                      simulate=simulate)
                )
            else:
                iptables['rules'].append(
                    utils.insert_rule('nat', kubesvc, 1,
                                      utils.make_rulespec(
                                          jump=kubesep
                                      ),
                                      utils.make_target_extensions(
                                          statistic=True,
                                          mode='nth',
                                          every=pod_num - i,
                                          packet=0
                                      ),
                                      simulate=simulate)
                )
        else:
            logging.error("Strategy Not Found! Use `random` or `roundrobin` Please")

        iptables['rules'].append(
            utils.insert_rule ('nat', kubesep, 1,
                              utils.make_rulespec(
                                  jump='DNAT',
                                  protocol=protocol,
                              ),
                              utils.make_target_extensions(
                                  match=protocol,
                                  to_destination=':'.join([pod_ip_list[i], str(target_port)])
                              ),
                              simulate=simulate)
        )
        iptables['rules'].append(
            utils.insert_rule('nat', kubesep, 1,
                              utils.make_rulespec(
                                  jump='KUBE-MARK-MASQ',
                                  source='/'.join([pod_ip_list[i], str(ip_prefix_len)])
                              ),
                              utils.make_target_extensions(),
                              simulate=simulate)
        )

    logging.info("Service [%s] Cluster IP: [%s] Port: [%s] TargetPort: [%s] Strategy: [%s]"
                 % (service_name, cluster_ip, port, target_port, strategy))


def save_iptables(path=default_iptables_path):
    """
    save current iptables to disk file, equals to the command:
    ```sudo iptables-save > path```
    :param path: saved file path
    :return: None
    """
    p = subprocess.Popen("iptables-save", stdout=subprocess.PIPE)
    f = open(path, "wb")
    f.write(p.stdout.read())
    f.close()
    logging.info("Save iptables successfully!")


def restore_iptables(path=default_iptables_path):
    """
    restore iptables from disk file, equals to the command:
    ```sudo iptables-restore < path```
    :param path: restored file path
    :return: None
    """
    f = open(path, "wb")
    p = subprocess.Popen("iptables-save", stdin=f)
    p.communicate()
    f.close()
    logging.info("Restore iptables successfully!")


def clear_iptables():
    """
    clear whole iptables, equals to the command:
    ```sudo iptables -F && sudo iptables -X```
    :return: None
    """
    utils.clear_rules()
    utils.dump_iptables()
    logging.info("Clear iptables successfully ...")


def example():
    init_iptables()
    set_iptables_clusterIP(cluster_ip='192.168.60.99',
                           service_name='example-service',
                           port=80,
                           target_port=80,
                           protocol='tcp',
                           pod_ip_list=['172.17.0.2', '172.17.0.3'],
                           strategy='random',
                           ip_prefix_len=32,
                           iptables=None)


if __name__ == '__main__':
    example()