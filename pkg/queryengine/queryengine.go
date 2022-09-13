package queryengine

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
)

func Run(ctx context.Context, wg *sync.WaitGroup, queryEnginePath, queryEnginePort, prismaSchemaFilePath string, enablePlayground bool) {
	args := []string{"--datamodel-path", prismaSchemaFilePath}
	if enablePlayground {
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
