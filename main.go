package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	r "reflect"
	"regexp"
	s "strings"
)

func resolveVolumeIdToDevPath(volumeID string) (string, error) {
	// TODO: handle error properly with descriptive error message instead of non-zero return code
	volID := s.Replace(volumeID, "-", "", 1)
	var data map[string]interface{}
	out, err := exec.Command("nvme", "list", "--output-format=json").Output()
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(out, &data); err != nil {
		return "", err
	}
	devices := data["Devices"].([]interface{})
	for _, device := range devices {
		d := device.(map[string]interface{})
		if d["SerialNumber"].(string) == volID {
			return d["DevicePath"].(string), err
		}
	}
	err = fmt.Errorf("device with volumeID %s is not present", volumeID)
	return "", err

}

func getInfoAboutVolume(volumeID string) (map[string]string, error) {
	devPath, err := resolveVolumeIdToDevPath(volumeID)
	res := make(map[string]string)
	if err != nil {
		return res, err
	}

	out, err := exec.Command("blkid", "--probe", "--output", "export", devPath).Output()
	if err != nil {
		return res, nil
	}
	outSlice := s.Split(string(out[:]), "\n")
	for _, info := range outSlice {
		if s.Contains(info, "=") {
			key, val := s.Split(info, "=")[0], s.Split(info, "=")[1]
			res[key] = val
		}
	}

	return res, nil
}

func createFs(volumeID string, mkfs string, label string, dryRun bool) error {
	devPath, err := resolveVolumeIdToDevPath(volumeID)
	if err != nil {
		return err
	}
	volumeInfo, err := getInfoAboutVolume(volumeID)
	if err != nil {
		return err
	}

	infoLabel := volumeInfo["LABEL"]
	infoType := volumeInfo["TYPE"]
	if infoLabel != "" && infoType != "" {
		if mkfs == infoType && label == infoLabel {
			return nil
		}
	}
	if dryRun {
		fmt.Printf("running command 'mkfs -t %s -L %s %s'\n", mkfs, label, devPath)
		return err
	}
	err = exec.Command("mkfs", "-t", mkfs, "-L", label, devPath).Run()
	return err
}

func writeFstab(newEntry []string, dryRun bool) error {
	file, err := os.Open("/etc/fstab")
	alreadyExists := false
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	b := make([]byte, fileInfo.Size())

	newFstabString := make([]string, 0)
	_, err = file.Read(b)
	if err != nil {
		return err
	}

	entries := s.Split(string(b), "\n")
	for _, entry := range entries {
		re := regexp.MustCompile("^(#|\\s|$)")
		reSpaces := regexp.MustCompile("\\s+")
		reEmptyLine := regexp.MustCompile("^\\s*$")
		if !re.Match([]byte(entry)) && r.DeepEqual(reSpaces.Split(string(entry), -1), newEntry) {
			alreadyExists = true
		}
		if !reEmptyLine.Match([]byte(entry)) {
			newFstabString = append(newFstabString, string(entry))
		}

	}
	err = file.Close()
	if err != nil {
		return err
	}

	if !alreadyExists {

		newFstabString = append(newFstabString, s.Join(newEntry, " "))

	}
	if dryRun {
		if !alreadyExists {
			fmt.Printf("writing to /etc/fstab entry: '%s'\n", s.Join(newEntry, " "))
		}
		return err
	}

	err = os.WriteFile("/etc/fstab", []byte(s.Join(newFstabString, "\n")), 0600)

	return err
}

func mount(mountpoint string, volumeID string, dryRun bool) error {
	devPath, err := resolveVolumeIdToDevPath(volumeID)
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("running command 'mount %s %s'\n", devPath, mountpoint)
		return err
	}
	err = os.MkdirAll(mountpoint, 0777)
	if err != nil {
		return err
	}
	err = exec.Command("mount", devPath, mountpoint).Run()
	return err

}
func main() {
	var mountpoint, label, mkfs, volumeID string
	var dryRun bool
	var cmd = &cobra.Command{
		Use: "ebsctl",
		Run: func(cmd *cobra.Command, args []string) {
			reVolume := regexp.MustCompile("vol-[a-z0-9]+")

			if !reVolume.Match([]byte(volumeID)) {
				fmt.Printf("%s doesn't match pattern vol-[a-z0-9]+\n", volumeID)
				return
			}
			if !(mkfs == "xfs" || mkfs == "ext4") {
				fmt.Printf("%s doesn't match patterns ['ext4', 'xfs']\n", mkfs)
				return
			}
			err := createFs(volumeID, mkfs, label, dryRun)
			if err != nil {
				log.Fatal(err)
			}
			err = mount(mountpoint, volumeID, dryRun)
			if err != nil {
				log.Fatal(err)
			}
			labelEntry := s.Join([]string{"LABEL=", label}, "")
			err = writeFstab([]string{labelEntry, mountpoint, mkfs, "defaults", "0", "0"}, dryRun)
			if err != nil {
				log.Fatal(err)
			}

		},
	}

	cmd.PersistentFlags().StringVar(&mountpoint, "mountpoint", "", "Mountpoint for the volume")
	cmd.PersistentFlags().StringVar(&label, "label", "", "Filesystem label")
	cmd.PersistentFlags().StringVar(&mkfs, "mkfs", "", "Filesystem type to create")
	cmd.PersistentFlags().StringVar(&volumeID, "volume-id", "", "Volume id")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Do not run program, but show the list of actions the tool will perform")
	cmd.MarkPersistentFlagRequired("mountpoint")
	cmd.MarkPersistentFlagRequired("label")
	cmd.MarkPersistentFlagRequired("mkfs")
	cmd.MarkPersistentFlagRequired("volume_id")
	cmd.Execute()
}
