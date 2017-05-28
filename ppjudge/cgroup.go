package main

import "os"
import "syscall"
import "io/ioutil"
import "strconv"
import "os/exec"
import "strings"
import "errors"

var cgroupDir = "/sys/fs/cgroup"

type Subsys struct {
	Sub   string
	Group string
}

func (ss *Subsys) convPath(name string) string {
	return cgroupDir + "/" + ss.Sub + "/" + ss.Group + "/" + name
}

func (ss *Subsys) listChildren() ([]string, error) {
	fis, err := ioutil.ReadDir(ss.convPath(""))

	if err != nil {
		return nil, err
	} else {
		groups := make([]string, 0, 10)

		for i := range fis {
			if fis[i].IsDir() {
				groups = append(groups, fis[i].Name())
			}
		}

		return groups, nil
	}
}

func (ss *Subsys) setVal(val, name string) error {
	fp, err := os.OpenFile(ss.convPath(name), syscall.O_WRONLY, 0664)

	if err != nil {
		return err
	}

	defer fp.Close()

	_, err = fp.Write([]byte(val))

	return err
}

func (ss *Subsys) setValInt(val int64, name string) error {
	return ss.setVal(strconv.FormatInt(val, 10), name)
}

func (ss *Subsys) addVal(val, name string) error {
	fp, err := os.OpenFile(ss.convPath(name), syscall.O_WRONLY|syscall.O_APPEND, 0664)

	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	defer fp.Close()

	_, err = fp.Write([]byte(val))

	return err
}

func (ss *Subsys) addValInt(val int64, name string) error {
	return ss.addVal(strconv.FormatInt(val, 10)+"\n", name)
}

func (ss *Subsys) getVal(name string) (*string, error) {
	fp, err := os.OpenFile(ss.convPath(name), syscall.O_RDONLY, 0664)

	if err != nil {
		return nil, err
	}

	defer fp.Close()

	data, err := ioutil.ReadAll(fp)

	if err != nil {
		return nil, err
	}

	str := string(data)
	return &str, nil
}

func (ss *Subsys) getValInt(name string) (int64, error) {
	res, err := ss.getVal(name)

	if err != nil {
		return 0, err
	}

	var str string
	if (*res)[len(*res)-1] == '\n' {
		str = (*res)[0 : len(*res)-1]
	} else {
		str = *res
	}

	return strconv.ParseInt(str, 10, 64)
}

type Cgroup struct {
	isCreated bool
	name      string
	subsys    []string
}

func (g *Cgroup) addSubsys(name string) *Cgroup {
	g.subsys = append(g.subsys, name)
	//g.libcg.AddController(name)

	return g
}

func (g *Cgroup) getSubsys(name string) *Subsys {
	return &Subsys{name, g.name}
}

func (g *Cgroup) Modify() error {
	if g.isCreated {
		return errors.New("Already created")
	}

	res, err := exec.Command("cgcreate", "-g", strings.Join(g.subsys, ",")+":/"+g.name).CombinedOutput()

	if len(res) != 0 {
		return errors.New("Failed to create a cgroup" + err.Error())
	}

	g.isCreated = true

	return nil
}

func (g *Cgroup) Delete() error {
	if !g.isCreated {
		return errors.New("Cgroup hasn't been created")
	}

	res, err := exec.Command("cgdelete", "-g", strings.Join(g.subsys, ",")+":/"+g.name).CombinedOutput()

	if len(res) != 0 {
		return errors.New("Failed to delete a cgroup " + err.Error())
	}

	return nil
}

// NewCgroup created a new Cgroup struct
func NewCgroup(name string) Cgroup {
	return Cgroup{false, name, []string{}}
}
