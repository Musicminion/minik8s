package gpu

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"miniK8s/pkg/k8log"
	"minik8s/apiObject/types"
	"minik8s/gpu/src/ssh"
	"minik8s/util/logger"
	"minik8s/util/recoverutil"
	"minik8s/util/uidutil"
	"minik8s/util/wait"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type JobArgs struct {
	JobName         string
	WorkDir         string
	Output          string
	Error           string
	NumProcess      int
	NumTasksPerNode int
	CpusPerTask     int
	GpuResources    string
	RunScripts      string
	CompileScripts  string
	Username        string
	Password        string
}

const pollPeriod = time.Minute * 5
const DefaultJobURL = "./usr/local/jobs"

type Server interface {
	Run()
}

type server struct {
	cli     ssh.Client
	args    JobArgs
	uid     types.UID
	jobsURL string
	jobID   string
}

func (s *server) recover() {
	if err := recover(); err != nil {
		fmt.Println(recoverutil.Trace(fmt.Sprintf("%v\n", err)))
		s.cli.Reconnect()
	}
}

func (s *server) poll() bool {
	defer s.recover()
	fmt.Println("Roll")
	return !s.cli.JobCompleted(s.jobID)
}

func (s *server) getCudaFiles() []string {
	var cudaFiles []string
	_ = filepath.WalkDir(s.jobsURL, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			fileName := d.Name()
			if strings.HasSuffix(fileName, ".cu") {
				cudaFiles = append(cudaFiles, fileName)
			}
		}
		return nil
	})
	fmt.Printf("cudaFiles: %v\n", cudaFiles)
	return cudaFiles
}

func (s *server) uploadSmallFiles(filenames []string) error {
	if resp, err := s.cli.Mkdir(s.args.WorkDir); err != nil {
		fmt.Println(resp)
		return err
	}
	for _, filename := range filenames {
		if file, err := os.Open(path.Join(s.jobsURL, filename)); err == nil {
			if content, err := ioutil.ReadAll(file); err == nil {
				_, _ = s.cli.WriteFile(path.Join(s.args.WorkDir, filename), string(content))
			}
		}
	}
	return nil
}

func (s *server) scriptPath() string {
	return path.Join(s.args.WorkDir, s.args.JobName+"-"+s.uid+".slurm")
}

func (s *server) createJobScript() error {
	template := `#!/bin/bash
#SBATCH --job-name=%s
#SBATCH --partition=dgx2
#SBATCH --output=%s
#SBATCH --error=%s
#SBATCH -N %d
#SBATCH --ntasks-per-node=%d
#SBATCH --cpus-per-task=%d
#SBATCH --gres=%s

%s
`
	script := fmt.Sprintf(
		template,
		s.args.JobName,
		s.args.Output,
		s.args.Error,
		s.args.NumProcess,
		s.args.NumTasksPerNode,
		s.args.CpusPerTask,
		s.args.GpuResources,
		strings.Replace(s.args.RunScripts, ";", "\n", -1),
	)
	_, err := s.cli.WriteFile(s.scriptPath(), script)
	return err
}

func (s *server) compile() error {
	_, err := s.cli.Compile(s.args.CompileScripts)
	return err
}

func (s *server) submitJob() (err error) {
	if s.jobID, err = s.cli.SubmitJob(s.scriptPath()); err == nil {
		fmt.Printf("submit succeed, got jod ID: %s\n", s.jobID)
	}
	return err
}

func (s *server) prepare() (err error) {
	cudaFiles := s.getCudaFiles()
	if len(cudaFiles) == 0 {
		return fmt.Errorf("no available cuda files")
	}
	if err = s.uploadSmallFiles(cudaFiles); err != nil {
		return err
	}
	fmt.Println("upload cuda files successfully")
	if err = s.compile(); err != nil {
		return err
	}
	fmt.Println("compile successfully")
	if err = s.createJobScript(); err != nil {
		return err
	}
	fmt.Println("create job script successfully")
	return nil
}

func (s *server) downloadResult() {
	outputFile := s.args.Output
	if content, err := s.cli.ReadFile(outputFile); err == nil {
		if file, err := os.Create(path.Join(s.jobsURL, outputFile)); err == nil {
			defer file.Close()
			_, _ = file.Write([]byte(content))
		}
	}

	errorFile := s.args.Error
	if content, err := s.cli.ReadFile(errorFile); err == nil {
		if file, err := os.Create(path.Join(s.jobsURL, errorFile)); err == nil {
			defer file.Close()
			_, _ = file.Write([]byte(content))
		}
	}
}

func (s *server) Run() {
	//在退出之前删除工作目录
	defer s.cli.RmDir(s.args.WorkDir)
	
	if err := s.prepare(); err != nil {
		k8log.ErrorLog("[Gpu]", "Prepare: failed to prepare")
		return
	}
	if err := s.submitJob(); err != nil {
		k8log.ErrorLog("[Gpu]", "Submit: failed to submit")
		return
	}
	//轮询任务状态,直到任务完成
	wait.PeriodWithCondition(pollPeriod, pollPeriod, s.poll)
	s.downloadResult()
	wait.Forever()
}

func NewServer(args JobArgs, jobsURL string) Server {
	return &server{
		cli:     ssh.NewClient(args.Username, args.Password),
		args:    args,
		uid:     uidutil.New(),
		jobsURL: jobsURL,
	}
}
