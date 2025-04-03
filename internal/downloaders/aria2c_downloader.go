package downloaders

import (
	"internal/common"
	"io"
	"os"
	"regexp"
)

type Aria2Downloader struct {
	DownloaderConfig *common.DownloaderConfig
}

func (d *Aria2Downloader) Download(url string, path string) error {
	l := common.NewLoggerWithPrefixAndColor("[Aria2Downloader.Download] ")
	l.Printf("url: %s, path: %s", url, path)

	// build arguments
	args := append([]string{}, d.DownloaderConfig.DefaultArgs...)

	// add customized args
	for _, argConf := range d.DownloaderConfig.Args {
		// TODO: Support more matcher types later.
		regex, err := regexp.Compile(argConf.Matcher.Pattern)
		if err != nil {
			l.Printf("failed to compile regexp: `%s`, error: %v", argConf.Matcher.Pattern, err)
			return err
		}
		if !regex.MatchString(url) {
			continue
		}
		args = append(args, argConf.Args...)
		l.Printf("matched pattern, adding args: %v", argConf.Args)
	}

	// replace placeholders
	for i, arg := range args {
		if arg == "$out" {
			args[i] = path
		}
		if arg == "$url" {
			args[i] = url
		}
	}

	l.Printf("got args: %s", args)

	// run commandline, and redirect output
	cmdline := d.DownloaderConfig.Cmd
	l.Printf("Run command: %s, %v", cmdline, args)
	err := common.RunCmd(cmdline, args, func(stdout io.ReadCloser) {
		io.Copy(os.Stdout, stdout)
	})
	if err != nil {
		l.Printf("failed to execute command `%s`, error: %v", cmdline, err)
		return err
	}

	return nil
}
