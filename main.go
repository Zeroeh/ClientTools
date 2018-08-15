package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"path/filepath"
)

const (
	menu string = "1. Check for updates.\n" +
		"2. Download client.\n" +
		"3. Export images.\n" +
		"4. Export binary data.\n" +
		"5. Add proxy server.\n" +
		"6. Change background.\n" +
		//"7. Update client.\n" +
		"> "
)

var (
	version string
	workingPath string
)

func checkErr(err error) {
	if err != nil {
		log.Println(err)
		checkMenu()
	}
}

func checkUpdates() {
	resp, err := http.Get("http://www.realmofthemadgod.com/version.txt")
	checkErr(err)
	defer resp.Body.Close()

	vers, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	localVers, err := ioutil.ReadFile(workingPath + "lib/version.txt")
	checkErr(err)

	if string(localVers) == string(vers) {
		fmt.Println("Game not updated, still on build", string(vers))
		checkMenu()
		return
	}

	version = string(vers)
	err = ioutil.WriteFile(workingPath + "lib/version.txt", vers, 0777)
	checkErr(err)

	fmt.Println("Game updated from ", string(localVers), "to", string(vers))
	downloadClient(true, true)
}

func downloadClient(update, menu bool) {
	if !update {
		resp, err := http.Get("http://www.realmofthemadgod.com/version.txt")
		checkErr(err)
		defer resp.Body.Close()

		version, err := ioutil.ReadAll(resp.Body)
		checkErr(err)

		err = ioutil.WriteFile(workingPath + "lib/version.txt", version, 0777)
		checkErr(err)
	}
	file, err := os.Create(workingPath + "client" + version + ".swf")
	checkErr(err)
	defer file.Close()

	resp, err := http.Get("http://www.realmofthemadgod.com/AssembleeGameClient" + string(version) + ".swf")
	checkErr(err)

	_, err = io.Copy(file, resp.Body)
	checkErr(err)

	fmt.Println("Client v" + string(version) + " saved.")
	if menu {
		checkMenu()
	}
}

func exportImages(menu bool) {
	if _, err := os.Stat(workingPath + "client" + version + ".swf"); os.IsNotExist(err) {
		downloadClient(false, false)
	}

	java, err := exec.LookPath("java")
	checkErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-export", "image", workingPath + "decompiled"+version+"/images", workingPath + "client"+version+".swf").Run()
	checkErr(err)

	files, err := ioutil.ReadDir(workingPath + "decompiled" + version + "/images")
	checkErr(err)

	r := regexp.MustCompile("([a-z.]+)(\\w+.jpg|\\w+.png)")

	for _, f := range files {
		if strings.Count(f.Name(), ".") > 1 {
			data, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/images/" + f.Name())
			checkErr(err)

			name := r.FindAllStringSubmatch(f.Name(), -1)
			path := strings.Replace(name[0][1], ".", "/", -1)
			os.MkdirAll(workingPath + "decompiled"+version+"/formatted/"+path, 0777)
			ioutil.WriteFile(workingPath + "decompiled"+version+"/formatted/"+path+name[0][2], data, 0777)
			fmt.Println(name[0][2], "saved.")
		}
	}

	fmt.Println("Images saved.")
	if menu {
		checkMenu()
	}
}

func exportBin() {
	if _, err := os.Stat(workingPath + "client" + version + ".swf"); os.IsNotExist(err) {
		downloadClient(false, false)
	}

	java, err := exec.LookPath("java")
	checkErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-export", "binaryData", workingPath + "decompiled"+version+"/binaryData", workingPath + "client"+version+".swf").Run()
	checkErr(err)

	files, err := ioutil.ReadDir(workingPath + "decompiled" + version + "/binaryData")
	checkErr(err)

	r := regexp.MustCompile("([a-z.]+)(\\w+.bin)")

	for _, f := range files {
		if strings.Count(f.Name(), ".") > 1 {
			data, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/binaryData/" + f.Name())
			checkErr(err)

			matches := r.FindAllStringSubmatch(f.Name(), -1)
			path := strings.Replace(matches[0][1], ".", "/", -1)
			name := strings.Replace(matches[0][2], ".bin", ".dat", -1)
			os.MkdirAll(workingPath + "decompiled"+version+"/formatted/"+path, 0777)
			ioutil.WriteFile(workingPath + "decompiled" + version + "/formatted/" + path + name, data, 0777)
			fmt.Println(name, "saved.")
		}
	}

	fmt.Println("Binary data saved.")
	checkMenu()
}

func updateClient() {

}

func addProxy() {
	if _, err := os.Stat(workingPath + "client" + version + ".swf"); os.IsNotExist(err) {
		downloadClient(false, false)
	}

	java, err := exec.LookPath("java")
	checkErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-selectclass", "kabam.rotmg.servers.control.ParseServerDataCommand", "-export", "script", workingPath + "decompiled"+version, workingPath + "client"+version+".swf").Run()
	checkErr(err)

	file, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/scripts/kabam/rotmg/servers/control/ParseServerDataCommand.as")
	checkErr(err)
	content := string(file)

	r := regexp.MustCompile("return _loc(\\d+)_;[\\s\\S]*?}")

	content = r.ReplaceAllString(content, "_loc${1}_.push(this.LocalhostServer());\n\t return _loc${1}_\n\t}\n\n\tpublic function LocalhostServer():Server\n\t{\n\treturn new Server().setName(\"## Proxy Server ##\").setAddress(\"127.0.0.1\").setPort(Parameters.PORT).setLatLong(Number(50000),Number(50000)).setUsage(0).setIsAdminOnly(false);\n\t}")

	ioutil.WriteFile(workingPath + "decompiled"+version+"/scripts/kabam/rotmg/servers/control/ParseServerDataCommand.as", []byte(content), 0644)

	err = exec.Command(java, "-jar", "ffdec.jar", "-replace", workingPath + "client"+version+".swf", workingPath + "client"+version+".swf", "kabam.rotmg.servers.control.ParseServerDataCommand", workingPath + "decompiled"+version+"/scripts/kabam/rotmg/servers/control/ParseServerDataCommand.as").Run()
	checkErr(err)

	fmt.Println("Proxy added.")
	checkMenu()
}

func changeBackground() {
	if _, err := os.Stat("background.png"); os.IsNotExist(err) {
		fmt.Println("background.png not found.")
		checkMenu()
	}

	if _, err := os.Stat(workingPath + "client" + version + ".swf"); os.IsNotExist(err) {
		downloadClient(false, false)
	}

	exportImages(false)

	java, err := exec.LookPath("java")
	checkErr(err)

	files, err := ioutil.ReadDir(workingPath + "decompiled" + version + "/images")
	checkErr(err)

	r := regexp.MustCompile("(\\d+)")

	for _, f := range files {
		if strings.Contains(f.Name(), "TitleView_TitleScreenGraphic") {
			matches := r.FindAllStringSubmatch(f.Name(), -1)
			err = exec.Command(java, "-jar", "ffdec.jar", "-replace", workingPath + "client"+version+".swf", workingPath + "client"+version+".swf", matches[0][1], workingPath + "background.png").Run()
			checkErr(err)
		}
	}

	fmt.Println("Background changed.")
	checkMenu()
}

func getWorkingModel(model int) {
	switch model {
	case 1:
		checkUpdates()
		return
	case 2:
		downloadClient(false, true)
		return
	case 3:
		exportImages(true)
		return
	case 4:
		exportBin()
		return
	case 5:
		addProxy()
		return
	case 6:
		changeBackground()
		return
	/*case 7:
	updateClient()
	return*/
	default:
		fmt.Print("Unknown model.")
	}
}

func checkMenu() {
	fmt.Print(menu)
	var menuInt int
	fmt.Scan(&menuInt)
	getWorkingModel(menuInt)
}

func main() {
	fmt.Println("Available", runtime.GOMAXPROCS(runtime.NumCPU()), "processes.")
	path, err := filepath.Abs("./")
	checkErr(err)
	workingPath = path + "/"
	vers, err := ioutil.ReadFile(workingPath + "lib/version.txt")
	checkErr(err)
	version = string(vers)
	checkMenu()
}
