package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nickalie/go-webpbin"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

func printLogo() {
	fmt.Print(`
                                 _            
                                | |           
  ____    ___  __      __  ___  | |__    _ __  
 / _  |  / _ \ \ \ /\ / / / _ \ | '_ \  | '_ \
| (_| | | (_) | \ V  V / |  __/ | |_) | | |_) |
\___, |  \___/   \_/\_/   \___| |_.__/  | .__/
 ___/ |                                 | |
|____/                                  |_|

`)
}


var (
	imageRe = regexp.MustCompile(`(?i)\.(jpe?g|png)$`)
	quality uint
	dir     string
	replace bool
	workers int
	dryRun bool
)

// set the flags
func init() {
	flag.StringVar(&dir, "d", "", "the directory to crawl")
	flag.UintVar(&quality, "q", 0, "the quality for the webp images")
	flag.BoolVar(&replace, "r", false, "replace existing webp files")
	flag.BoolVar(&dryRun, "dry-run", false, "whether to handle this as a dry run and only " +
		"print target files")
	flag.IntVar(&workers, "w", runtime.NumCPU(), "the number of worker routines to spawn. " +
		"Defaults to number of CPUs.")
}


func mustGetFileSize(file string) int64 {
	fi, err := os.Stat(file)
	if err != nil {
		panic(err)
	}
	return fi.Size()
}

type webpJobResult struct {
	err         error
	compression float64
	exists      bool
	outputFile  string
}

func newJob(input string, quality uint) *job {
	j := &job{
		input:   input,
		quality: quality,
		resCh:   make(chan *webpJobResult),
	}
	return j
}

type job struct {
	input   string
	quality uint
	res     *webpJobResult
	resCh   chan *webpJobResult
}

// waitForResult gets a result for this job only when job completion signal is set
func (j *job) waitForResult() *webpJobResult {
	j.res = <-j.resCh
	return j.res
}

func newPool(ctx context.Context, workers int) *pool {
	ctx, done := context.WithCancel(ctx)
	p := &pool{
		workers: workers,
		jobs:    make(chan *job),
		ctx:     ctx,
		done:    done,
		wg:      &sync.WaitGroup{},
	}
	p.start()
	return p
}

type pool struct {
	workers int
	jobs    chan *job
	ctx     context.Context
	done    context.CancelFunc
	wg      *sync.WaitGroup
}

// execute executes a compression job
func (p *pool) execute(j *job) {
	go j.waitForResult()
	r := &webpJobResult{}

	// always pass the result to the job's result channel
	defer func() {
		j.resCh <- r
		close(j.resCh)
	}()

	var (
		targetExt string
	)

	// get the absolute path
	j.input, r.err = filepath.Abs(j.input)
	if r.err != nil {
		return
	}

	// get the target's extension
	targetExt = filepath.Ext(j.input)

	// output is the old filepath with new webp extension
	r.outputFile = j.input[:len(j.input)-len(targetExt)] + ".webp"

	// check if file already exists
	if !replace {
		if _, err := os.Stat(r.outputFile); err == nil {
			// file already exists
			r.exists = true
			log.Println(j.input, "already has a webp version")
			return
		}
	}

	if dryRun {
		// if it's a dry run then just print and return
		log.Printf("%s \u2192 %s [?]\n", j.input, r.outputFile)
		return
	}


	// get the size of the original file
	fSizeTarget := mustGetFileSize(j.input)

	r.err = webpbin.NewCWebP().
		Quality(j.quality).
		InputFile(j.input).
		OutputFile(r.outputFile).
		Run()

	if r.err != nil {
		return
	}

	// get the file size of the new file
	fSizeOutput := mustGetFileSize(r.outputFile)

	// calculate the compression percentage
	r.compression = (1 - (float64(fSizeOutput) / float64(fSizeTarget))) * 100

	if r.err != nil {
		log.Printf("!ERROR webp generation for %s FAILED with error: %s\n", r.err)
	} else {
		log.Printf("%s \u2192 %s [%.2f%%]\n", j.input, r.outputFile, r.compression)
	}

	return
}

func (p *pool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *pool) wait() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *pool) stop() {
	p.done()
	p.wg.Wait()
}

func (p *pool) worker() {
	defer func() {
		p.wg.Done()
	}()
	for {
		select {
		case j, ok := <-p.jobs:
			if !ok {
				// no more work
				return
			}
			// execute a job and pass the result into the result channel
			p.execute(j)
		case <-p.ctx.Done():
			// we'imageRe done early
			return
		}
	}
}

func main() {
	printLogo()
	flag.Parse()
	if (len(dir) < 1 || quality < 1) && !dryRun {
		// print help
		fmt.Print(`
gowebp is a tool used to create webp images from jpegs and png files

Usage:
`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	p := newPool(context.Background(), workers)

	dir = strings.TrimSpace(dir)

	dir, err := filepath.Abs(dir)

	if err != nil {
		fmt.Println("dir is not valid!")
		os.Exit(2)
	}

	fmt.Println("CRAWLING:\t", dir)
	fmt.Println("QUALITY:\t", quality)
	fmt.Println("WORKERS:\t", workers)
	if dryRun {
		fmt.Println("*** THIS IS A DRY RUN ***")
	}

	// stop pool when exiting
	defer p.stop()

	cnt := 0
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if imageRe.MatchString(info.Name()) {
				//log.Println("found image:", path)
				p.jobs <- newJob(path, quality)
				cnt += 1
			}

			return nil
		})
	if err != nil {
		log.Println("!!ERROR", err)
	}

	p.wait()
}
