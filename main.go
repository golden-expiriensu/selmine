package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/sync/errgroup"
)

const prefix = "claim"

type Worker struct {
	state crypto.KeccakState
}

func (w Worker) work(ctx context.Context, aim [4]byte, start int, max int) int {
	for {
		select {
		case <-ctx.Done():
			return -1
		default:
			hash := w.calc(start)
			if hash[0] == aim[0] &&
				hash[1] == aim[1] &&
				hash[2] == aim[2] &&
				hash[3] == aim[3] {
				return start
			}
			start++
			if start > max {
				return -1
			}
		}
	}
}

func (w Worker) calc(salt int) common.Hash {
	input := fmt.Sprintf("%s%d()", prefix, salt)
	return crypto.HashData(w.state, []byte(input))
}

func main() {
	var desired = [4]byte{0x5a, 0x7b, 0x87, 0xf2}
	//var desired = [4]byte{0x1c, 0xcc, 0x8c, 0x50}
	fmt.Printf("Desired selector is %#x\n", desired)
	fmt.Printf("Template is %sXXX()\n", prefix)

	workers := make([]Worker, 0)
	result := -1
	wg, ctx := errgroup.WithContext(context.Background())

	const start = 800000008
	const step = 100000001

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Go(func() error {
			worker := Worker{crypto.NewKeccakState()}
			workers = append(workers, worker)
			s := start + i*step
			fmt.Printf("Started worker %d from %d to %d\n", i, s, s+step)
			res := worker.work(ctx, desired, s, s+step-1)
			if res != -1 {
				result = res
				return errors.New("Done")
			}
			return nil
		})
	}

	_ = wg.Wait()

	fmt.Printf("Result is %d", result)
}
