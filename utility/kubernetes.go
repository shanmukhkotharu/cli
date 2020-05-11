package utility

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func ObtainKubeConfig(localKubeconfig string, civoConfig string, merge bool) error {

	kubeconfig := []byte(civoConfig)

	if merge {
		// Create a merged kubeconfig
		kubeconfig, _ = mergeConfigs(localKubeconfig, []byte(kubeconfig))
	}

	// Create a new kubeconfig
	if writeErr := writeConfig(localKubeconfig, []byte(kubeconfig), false); writeErr != nil {
		return writeErr
	}
	return nil
}

func mergeConfigs(localKubeconfigPath string, k3sconfig []byte) ([]byte, error) {
	// Create a temporary kubeconfig to store the config of the newly create k3s cluster
	file, err := ioutil.TempFile(os.TempDir(), "civo-temp-*")
	if err != nil {
		return nil, fmt.Errorf("could not generate a temporary file to store the kuebeconfig: %s", err)
	}
	defer file.Close()

	if writeErr := writeConfig(file.Name(), []byte(k3sconfig), true); writeErr != nil {
		return nil, writeErr
	}

	fmt.Printf("Merging with existing kubeconfig at %s\n", localKubeconfigPath)

	// Append KUBECONFIGS in ENV Vars
	appendKubeConfigENV := fmt.Sprintf("KUBECONFIG=%s:%s", localKubeconfigPath, file.Name())

	// Merge the two kubeconfigs and read the output into 'data'
	cmd := exec.Command("kubectl", "config", "view", "--merge", "--flatten")
	cmd.Env = append(os.Environ(), appendKubeConfigENV)
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not merge kubeconfigs: %s", err)
	}

	// Remove the temporarily generated file
	err = os.Remove(file.Name())
	if err != nil {
		return nil, fmt.Errorf("could not remove temporary kubeconfig file: %s", file.Name())
	}

	return data, nil
}

// Generates config files give the path to file: string and the data: []byte
func writeConfig(path string, data []byte, suppressMessage bool) error {
	if !suppressMessage {
		fmt.Printf("Saving file to: %s\n", path)
		fmt.Printf("\n# Test your cluster with:\nexport KUBECONFIG=%s\nkubectl get node -o wide\n", path)
	}

	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer file.Close()
	}

	writeErr := ioutil.WriteFile(path, []byte(data), 0600)
	if writeErr != nil {
		return writeErr
	}
	return nil
}