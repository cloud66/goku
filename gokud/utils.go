package main

import (
	"bufio"
	"errors"
	"os/exec"
	"io"
	"os"
	"strconv"
	"strings"
)

func lookupGroupId(group string) (gid int, err error) {
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

func lookupUserId(user string) (uid int, err error) {
	cmd, err := exec.Command("id","-u", user).Output()
	if err != nil {
		return -1, errors.New("user not found")
	}

	result, err := strconv.Atoi(strings.Trim(string(cmd), "\n"))
	if err != nil {
		return -1, err
	}

	return result, nil
}
