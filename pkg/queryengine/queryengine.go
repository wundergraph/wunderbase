package queryengine

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
)

func Run(ctx context.Context, wg *sync.WaitGroup, queryEnginePath, queryEnginePort, prismaSchemaFilePath string, production bool) {
	// when start prisma query engine ,
	// we're not able to listen on the same port,
	// if last engine instance still alive.
	// so we must kill the existing engine process before we start new onw.

	args := []string{"--datamodel-path", prismaSchemaFilePath}
	if !production {
		killExistingPrismaQueryEngineProcess(queryEnginePort)
		args = append(args, "--enable-playground", "--port", queryEnginePort)
	}
	cmd := exec.CommandContext(ctx, queryEnginePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalln("run query engine", err)
	}
	<-ctx.Done()
	err = cmd.Process.Kill()
	if err != nil {
		log.Println("kill query engine", err)
	}
	log.Println("Query Engine stopped")
	wg.Done()
}

// reference:https://github.com/wundergraph/wundergraph
func killExistingPrismaQueryEngineProcess(queryEnginePort string) {
	if runtime.GOOS == "windows" {
		command := fmt.Sprintf("(Get-NetTCPConnection -LocalPort %s).OwningProcess -Force", queryEnginePort)
		execCmd(exec.Command("Stop-Process", "-Id", command))
	} else {
		command := fmt.Sprintf("lsof -i tcp:%s | grep LISTEN | awk '{print $2}' | xargs kill -9", queryEnginePort)
		execCmd(exec.Command("bash", "-c", command))
	}
}

func execCmd(cmd *exec.Cmd) {
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			log.Println("Error during port killing (exit code: )", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
		}
	} else {
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)

		log.Println("Successfully killed existing prisma query process", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
	}
}
