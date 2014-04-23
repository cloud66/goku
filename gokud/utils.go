package main

import (
	"bufio"
	"errors"
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
	f, err := os.Open("id -u " + user)
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
		if len(p) >= 3 && p[0] == user {
			return strconv.Atoi(p[2])
		}
	}
	return 0, errors.New("user not found")
}
