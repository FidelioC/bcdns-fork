package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	COM_TLD_SPECS  = "../all_specs/com_tldSpec.json"
	EXAMPLE_SPECS  = "../all_specs/exampleSpec.json"
	WHATEVER_SPECS = "../all_specs/whateverSpec.json"
)

var (
	FILLER_TLDS        = [...]string{"org", "net", "gov", "edu"}
	FILLER_TARGETS     = [...]string{"test", "example2", "example3", "domain", "domain1", "domain2", "domain3", "google", "facebook", "twitter", "instagram", "linkedin", "youtube", "reddit", "tiktok", "snapchat", "whatsapp", "telegram", "signal", "discord", "slack", "microsoft", "apple", "amazon", "netflix", "spotify", "uber", "lyft", "airbnb", "expedia", "tripadvisor", "booking", "priceline", "kayak"}
	DNS_CONFIGURATIONS = [...]ArchOptions{
		{
			tlds:         1,
			networks:     2,
			validators:   2,
			normal_nodes: 2,
		},
	}
	currentConfig *ArchOptions
)

type ArchOptions struct {
	tlds         int
	networks     int
	validators   int
	normal_nodes int
}

type ArchOption func(*ArchOptions)

func withOptions(options *ArchOptions) ArchOption {
	return func(o *ArchOptions) {
		o.tlds = options.tlds
		o.networks = options.networks
		o.validators = options.validators
		o.normal_nodes = options.normal_nodes
	}
}

func runCommand(cmd *exec.Cmd, name string) error {
	fmt.Printf("[CMD] Starting: %s\n", name)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}

	// stream output
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("[%s] %s\n", name, scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[%s ERR] %s\n", name, scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	fmt.Printf("[CMD] Completed: %s\n", name)
	return nil
}

func launchArch(setters ...ArchOption) {
	o := &ArchOptions{}

	for _, setter := range setters {
		setter(o)
	}

	cmd := exec.Command("./launch_dns_arch.sh",
		"--tld",
		strconv.Itoa(o.tlds),
		"--target",
		strconv.Itoa(o.networks),
		"--validators",
		strconv.Itoa(o.validators),
		"--nodes",
		strconv.Itoa(o.normal_nodes))
	cmd.Dir = "../"

	if err := runCommand(cmd, "launch_dns_arch.sh"); err != nil {
		fmt.Printf("Error executing launch_dns_arch.sh: %v\n", err)
	} else {
		fmt.Println("Successfully launched DNS architecture.")
	}
}

func cleanArch() {
	cmd := exec.Command("./dns_arch_cleanup.sh")
	cmd.Dir = "../"

	if err := runCommand(cmd, "dns_arch_cleanup.sh"); err != nil {
		fmt.Printf("Error executing dns_arch_cleanup.sh: %v\n", err)
	} else {
		fmt.Println("Successfully cleaned up DNS architecture.")
	}
}

func copySHFiles() error {
	files, err := filepath.Glob("*.sh")
	if err != nil {
		return err
	}

	for _, file := range files {
		src, err := os.Open(file)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join("..", file))
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}

		err = os.Chmod(filepath.Join("..", file), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func setupDNSInfo() {
	fmt.Println("Setting up DNS information...")

	// npm install
	cmd := exec.Command("npm", "install")
	cmd.Dir = "../../dns_client"
	if err := runCommand(cmd, "npm install"); err != nil {
		fmt.Printf("Error during npm install: %v\n", err)
	}

	// Register COM TLD
	cmd = exec.Command("npm", "run", "register", "--",
		"--tld", "com",
		"../polkadot-sdk-solochain-template/all_specs/com_tldSpec.json",
		"//Alice")
	cmd.Dir = "../../dns_client"
	if err := runCommand(cmd, "register COM TLD"); err != nil {
		fmt.Printf("Error registering COM TLD: %v\n", err)
	}

	// Register filler TLDs
	for _, tld := range FILLER_TLDS {
		cmd := exec.Command("npm", "run", "register", "--",
			"--tld", tld,
			"../polkadot-sdk-solochain-template/all_specs/com_tldSpec.json",
			"//Alice")
		cmd.Dir = "../../dns_client"

		if err := runCommand(cmd, "register "+tld+" TLD"); err != nil {
			fmt.Printf("Error registering %s TLD: %v\n", tld, err)
		}
	}

	time.Sleep(time.Millisecond * 10000) // Wait 10 seconds

	// Register example.com
	cmd = exec.Command("npm", "run", "register", "--",
		"--domain", "example.com",
		"../polkadot-sdk-solochain-template/all_specs/exampleSpec.json",
		"//Alice")
	cmd.Dir = "../../dns_client"
	if err := runCommand(cmd, "register example.com"); err != nil {
		fmt.Printf("Error registering example.com: %v\n", err)
	}

	// Register whatever.com
	cmd = exec.Command("npm", "run", "register", "--",
		"--domain", "whatever.com",
		"../polkadot-sdk-solochain-template/all_specs/whateverSpec.json",
		"//Alice")
	cmd.Dir = "../../dns_client"
	if err := runCommand(cmd, "register whatever.com"); err != nil {
		fmt.Printf("Error registering whatever.com: %v\n", err)
	}

	// Register filler targets
	for _, target := range FILLER_TARGETS {
		cmd := exec.Command("npm", "run", "register", "--",
			"--domain", target+".com",
			"../polkadot-sdk-solochain-template/all_specs/exampleSpec.json",
			"//Alice")
		cmd.Dir = "../../dns_client"

		if err := runCommand(cmd, "register "+target+".com"); err != nil {
			fmt.Printf("Error registering %s.com: %v\n", target, err)
		}
	}

	time.Sleep(time.Millisecond * 10000) // Wait 10 seconds
}

func main() {
	flagInit := flag.Bool("init", false, "Initialize architecture and setup DNS")
	flagCleanup := flag.Bool("cleanup", false, "Clean up architecture")
	flag.Parse()

	if err := copySHFiles(); err != nil {
		fmt.Printf("Error copying .sh files: %v\n", err)
	}

	for _, o := range DNS_CONFIGURATIONS {
		currentConfig = &o

		switch {
		case *flagInit:
			fmt.Printf("Launching architecture using settings:\n %+v\n", o)
			launchArch(withOptions(&o))
			setupDNSInfo()
		case *flagCleanup:
			fmt.Printf("Cleaning up architecture using settings:\n %+v\n", o)
			cleanArch()
		default:
			fmt.Printf("Launching architecture using settings:\n %+v\n", o)
			launchArch(withOptions(&o))
			setupDNSInfo()
			cleanArch()
		}
	}
}
