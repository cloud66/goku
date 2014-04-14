package main

import (
	"bufio"
	"io"
	"os"
	"errors"
	"strconv"
	"strings"
)

func LookupGroupId(group string) (gid int, err error) {
	f, err := os.Open("/etc/group")
	if err != nil {
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)
	for {
		s, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		p := strings.Split(s, ":")
		if len(p) >= 3 && p[0] == group {
			return strconv.Atoi(p[2])
		}
	}
	return 0, errors.New("group not found")
}
