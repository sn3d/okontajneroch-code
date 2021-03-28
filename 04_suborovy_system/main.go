package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		panic("run?")
	}

	switch os.Args[1] {
	case "run":
		run()
	case "reexec":
		reexec()
	default:
		panic("run?")
	}
}

func run() {
	// pripravim si re-exec
	cmd := exec.Command("/proc/self/exe", append([]string{"reexec"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// nastavim menne priestory pre re-exec
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,

		//mapovanie použivateľa
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},

		// mapovanie skupiny
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
	}

	// reexec
	cmd.Run()
}

func reexec() {
	// urobim potrebnu inicializaciu

	// vymenime suborovy system
	changeFs("rootfs")

	// zmodifikujem si prompt, aby som vedel rozlíšiť
	// či som v hosťovskom systéme alebo v behovom protredí
	os.Setenv("PS1", "anton> ")

	// pripravim a spustim podproces
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// funkcia prehodi aktualny suboroby system
func changeFs(rootfs string) error {
	_, err := os.Stat(rootfs)
	if os.IsNotExist(err) {
		panic("chyba rootfs adresar")
	}

	// pripravime proc v rootfs
	target := filepath.Join(rootfs, "proc")
	err = os.MkdirAll(target, 0755)
	if err != nil {
		return err
	}

	// kedze sme v novom namespace, do rootfs
	// bude namontovany novy proc
	err = syscall.Mount("proc", target, "proc", uintptr(0), "")
	if err != nil {
		return err
	}

	// tento 'hack' je potrebny, lebo pivot_root nevie
	// prehodit obycajny adresar
	err = syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return err
	}

	// vytvorime docastny adresar, kde bude umiestneny stary
	// suborovy system
	old := filepath.Join(rootfs, "tmp", "old")
	err = os.MkdirAll(old, 0700)
	if err != nil {
		return err
	}

	// pivotovanie
	err = syscall.PivotRoot(rootfs, old)
	if err != nil {
		return err
	}

	// vyskocime do noveho systemu
	err = os.Chdir("/")
	if err != nil {
		return err
	}

	// odmontujeme stary suborovy system
	err = syscall.Unmount("/tmp/old", syscall.MNT_DETACH)
	if err != nil {
		return err
	}

	// zmazeme docasny adresar
	err = os.RemoveAll("/tmp/old")
	if err != nil {
		return err
	}

	return nil
}
