package migrate

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"time"
)

type MigrationRequest struct {
	Id      int                    `json:"id"`
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  MigrationRequestParams `json:"params"`
}

type MigrationRequestParams struct {
	Force  bool   `json:"force"`
	Schema string `json:"schema"`
}

type MigrationResponse struct {
	Jsonrpc string                   `json:"jsonrpc"`
	Result  *MigrationResponseResult `json:"result,omitempty"`
	Error   *MigrationResponseError  `json:"error,omitempty"`
}

type MigrationResponseResult struct {
	ExecutedSteps int `json:"executedSteps"`
}

type MigrationResponseError struct {
	Code    int                        `json:"code"`
	Message string                     `json:"message"`
	Data    MigrationResponseErrorData `json:"data"`
}

type MigrationResponseErrorData struct {
	IsPanic bool                           `json:"is_panic"`
	Message string                         `json:"message"`
	Meta    MigrationResponseErrorDataMeta `json:"meta"`
}

type MigrationResponseErrorDataMeta struct {
	FullError string `json:"full_error"`
}

func Database(migrationEnginePath, migrationLockFilePath, schema, schemaPath string) {

	h := sha256.New()
	expected := h.Sum([]byte(schema))
	lock, _ := ioutil.ReadFile(migrationLockFilePath)
	if bytes.Equal(lock, expected) {
		log.Println("Migration already executed, skipping")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cmd := exec.CommandContext(ctx, migrationEnginePath, "--datamodel", schemaPath)
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln("migration engine std in pipe", err)
	}
	defer in.Close()

	req := MigrationRequest{
		Id:      1,
		Jsonrpc: "2.0",
		Method:  "schemaPush",
		Params: MigrationRequestParams{
			Force:  true,
			Schema: schema,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		log.Fatalln("marshal migration request", err)
	}
	data = append(data, []byte("\n")...)
	_, err = in.Write(data)

	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("migration std out pipe", err)
	}

	go func() {
		r := bufio.NewReader(out)
		outBuf := &bytes.Buffer{}
		for {
			b, err := r.ReadByte()
			if err != nil {
				log.Fatalln("migration ReadByte", err)
			}
			err = outBuf.WriteByte(b)
			if err != nil {
				log.Fatalln("migration writeByte", err)
			}
			if b == '\n' {
				cancel()
				var resp MigrationResponse
				err = json.Unmarshal(outBuf.Bytes(), &resp)
				if err != nil {
					log.Fatalln("migration unmarshal response", err)
				}
				if resp.Error == nil {
					log.Println("Migration successful, updating lock file")
					err = ioutil.WriteFile(migrationLockFilePath, expected, 0644)
					if err != nil {
						log.Fatalln("migration write lock file", err)
					}
					return
				}
				pretty, err := json.MarshalIndent(resp, "", "  ")
				if err != nil {
					log.Fatalln("migration marshal error", err)
				}
				log.Printf("Migration failed:\n%s", string(pretty))
				return
			}
		}
	}()

	err = cmd.Run()
	if err != nil && ctx.Err() == nil {
		log.Println("migration engine run", err)
		err = nil
	}
	err = ioutil.WriteFile(migrationLockFilePath, expected, 0644)
	if err != nil {
		log.Fatalln("migration write lock file", err)
	}
}
