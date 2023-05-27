import os
import func
import json

from flask import Flask, request

app = Flask(__name__)
host_conf = "0.0.0.0"
port_conf = os.environ.get('PORT', 18080)


@app.route('/', methods=['GET'])
def welecome():
    welecomeWords = "Welcome to use this server!"
    usage = "Usage: send put/post request to this url!"
    return welecomeWords + '\n' + usage

@app.route("/", methods=['POST'])
def callCloudFuncByPost():
    try:
        userparams = json.loads(request.get_data())
    except json.JSONDecodeError:
        userparams = ""
    finally:
        res = func.main(userparams)
    return json.dumps(res)

@app.route("/", methods=['PUT'])
def callCloudFuncByPut():
    try:
        userparams = json.loads(request.get_data())
    except json.JSONDecodeError:
        userparams = ""
    finally:
        res = func.main(userparams)
    return json.dumps(res)

# @app.route("/test", methods=['GET'])
# def test():
#     userparams = "json.loads(request.get_data())"
#     res = func.main(userparams)
#     return json.dumps(res)

@app.route("/config", methods=['GET'])
def getConfig():
    config = {
        "host": host_conf,
        "port": port_conf,
    }
    return json.dumps(config)

if __name__ == '__main__':
    app.run(host=host_conf, port=port_conf, debug=False)