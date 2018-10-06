package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	menu string = "1. Check for updates.\n" +
		"2. Download client.\n" +
		"3. Export images.\n" +
		"4. Export binary data.\n" +
		"5. Add proxy server.\n" +
		"6. Change background.\n" +
		"7. Export packets.\n" +
		"8. Export assets config.\n" +
		"> "
)

var (
	versionURL = "https://realmofthemadgodhrd.appspot.com/version.txt"
	clientURL = "https://realmofthemadgodhrd.appspot.com/AssembleeGameClient"
	version     string
	workingPath string
)

func logErr(err error) {
	if err != nil {
		log.Println(err)
		checkMenu()
	}
}

func checkUpdates() {
	resp, err := http.Get(versionURL)
	logErr(err)
	defer resp.Body.Close()

	vers, err := ioutil.ReadAll(resp.Body)
	logErr(err)

	localVers, err := ioutil.ReadFile(workingPath + "lib/version.txt")
	logErr(err)

	if string(localVers) == string(vers) {
		fmt.Println("Game not updated, still on build", string(vers))
		checkMenu()
		return
	}

	version = string(vers)
	err = ioutil.WriteFile(workingPath+"lib/version.txt", vers, 0666)
	logErr(err)

	fmt.Println("Game updated from ", string(localVers), "to", string(vers))
	downloadClient(true, true)
}

func downloadClient(update, menu bool) {
	if !update {
		resp, err := http.Get(versionURL)
		logErr(err)
		defer resp.Body.Close()

		version, err := ioutil.ReadAll(resp.Body)
		logErr(err)

		err = ioutil.WriteFile(workingPath+"lib/version.txt", version, 0666)
		logErr(err)
	}
	file, err := os.Create(workingPath + "client" + version + ".swf")
	logErr(err)
	defer file.Close()

	resp, err := http.Get(clientURL + string(version) + ".swf")
	logErr(err)

	_, err = io.Copy(file, resp.Body)
	logErr(err)

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
	logErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-export", "image", workingPath+"decompiled"+version+"/images", workingPath+"client"+version+".swf").Run()
	logErr(err)

	files, err := ioutil.ReadDir(workingPath + "decompiled" + version + "/images")
	logErr(err)

	r := regexp.MustCompile("([a-z.]+)(\\w+.jpg|\\w+.png)")

	for _, f := range files {
		if strings.Count(f.Name(), ".") > 1 {
			data, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/images/" + f.Name())
			logErr(err)

			name := r.FindAllStringSubmatch(f.Name(), -1)
			path := strings.Replace(name[0][1], ".", "/", -1)
			os.MkdirAll(workingPath+"decompiled"+version+"/formatted/"+path, 0666)
			ioutil.WriteFile(workingPath+"decompiled"+version+"/formatted/"+path+name[0][2], data, 0666)
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
	logErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-export", "binaryData", workingPath+"decompiled"+version+"/binaryData", workingPath+"client"+version+".swf").Run()
	logErr(err)

	files, err := ioutil.ReadDir(workingPath + "decompiled" + version + "/binaryData")
	logErr(err)

	r := regexp.MustCompile("([a-z.]+)(\\w+.bin)")

	for _, f := range files {
		if strings.Count(f.Name(), ".") > 1 {
			data, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/binaryData/" + f.Name())
			logErr(err)

			matches := r.FindAllStringSubmatch(f.Name(), -1)
			path := strings.Replace(matches[0][1], ".", "/", -1)
			name := strings.Replace(matches[0][2], ".bin", ".dat", -1)
			os.MkdirAll(workingPath+"decompiled"+version+"/formatted/"+path, 0666)
			ioutil.WriteFile(workingPath+"decompiled"+version+"/formatted/"+path+name, data, 0666)
			fmt.Println(name, "saved.")
		}
	}

	fmt.Println("Binary data saved.")
	checkMenu()
}

func addProxy() {
	if _, err := os.Stat(workingPath + "client" + version + ".swf"); os.IsNotExist(err) {
		downloadClient(false, false)
	}

	java, err := exec.LookPath("java")
	logErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-selectclass", "kabam.rotmg.servers.control.ParseServerDataCommand", "-export", "script", workingPath+"decompiled"+version, workingPath+"client"+version+".swf").Run()
	logErr(err)

	file, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/scripts/kabam/rotmg/servers/control/ParseServerDataCommand.as")
	logErr(err)
	content := string(file)

	r := regexp.MustCompile("return _loc(\\d+)_;[\\s\\S]*?}")

	content = r.ReplaceAllString(content, "_loc${1}_.push(this.LocalhostServer());\n\t return _loc${1}_\n\t}\n\n\tpublic function LocalhostServer():Server\n\t{\n\treturn new Server().setName(\"## Proxy Server ##\").setAddress(\"127.0.0.1\").setPort(Parameters.PORT).setLatLong(Number(50000),Number(50000)).setUsage(0).setIsAdminOnly(false);\n\t}")

	ioutil.WriteFile(workingPath+"decompiled"+version+"/scripts/kabam/rotmg/servers/control/ParseServerDataCommand.as", []byte(content), 0666)

	err = exec.Command(java, "-jar", "ffdec.jar", "-replace", workingPath+"client"+version+".swf", workingPath+"client"+version+".swf", "kabam.rotmg.servers.control.ParseServerDataCommand", workingPath+"decompiled"+version+"/scripts/kabam/rotmg/servers/control/ParseServerDataCommand.as").Run()
	logErr(err)

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
	logErr(err)

	files, err := ioutil.ReadDir(workingPath + "decompiled" + version + "/images")
	logErr(err)

	r := regexp.MustCompile("(\\d+)")

	for _, f := range files {
		if strings.Contains(f.Name(), "TitleView_TitleScreenGraphic") {
			matches := r.FindAllStringSubmatch(f.Name(), -1)
			err = exec.Command(java, "-jar", "ffdec.jar", "-replace", workingPath+"client"+version+".swf", workingPath+"client"+version+".swf", matches[0][1], workingPath+"background.png").Run()
			logErr(err)
		}
	}

	fmt.Println("Background changed.")
	checkMenu()
}

func exportPackets() {
	if _, err := os.Stat(workingPath + "client" + version + ".swf"); os.IsNotExist(err) {
		downloadClient(false, false)
	}

	java, err := exec.LookPath("java")
	logErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-selectclass", "kabam.rotmg.messaging.impl.GameServerConnection", "-export", "script", workingPath+"decompiled"+version, workingPath+"client"+version+".swf").Run()
	logErr(err)

	gsc, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/scripts/kabam/rotmg/messaging/impl/GameServerConnection.as")
	logErr(err)
	content := string(gsc)

	r, err := regexp.Compile("const ([\\s\\S]*?):int[\\s\\S]*?(\\d+);")
	logErr(err)
	matches := r.FindAllStringSubmatch(content, -1)

	as, err := os.Create(workingPath + "decompiled" + version + "/AS3.as")
	logErr(err)
	defer as.Close()
	xml, err := os.Create(workingPath + "decompiled" + version + "/K-Relay.xml")
	logErr(err)
	defer xml.Close()

	xml.WriteString("<Packets>\n")
	for i := 0; i < len(matches); i++ {
		as.WriteString("public static const " + matches[i][1] + ":int = " + matches[i][2] + ";\n")
		xml.WriteString("	<Packet>\n		<PacketName>" + strings.Replace(matches[i][1], "_", "", -1) + "</PacketName>\n		<PacketID>" + matches[i][2] + "</PacketID>\n	</Packet>\n")
	}
	xml.WriteString("</Packets>")

	fmt.Println("Packets saved.")
	checkMenu()
}

func exportAssetsConfig() {
	if _, err := os.Stat(workingPath + "client" + version + ".swf"); os.IsNotExist(err) {
		downloadClient(false, false)
	}

	java, err := exec.LookPath("java")
	logErr(err)

	err = exec.Command(java, "-jar", "ffdec.jar", "-selectclass", "com.company.assembleegameclient.util.AssetLoader", "-export", "script", workingPath+"decompiled"+version, workingPath+"client"+version+".swf").Run()
	logErr(err)

	assetLoader, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/scripts/com/company/assembleegameclient/util/AssetLoader.as")
	logErr(err)

	ioutil.WriteFile(workingPath+"decompiled"+version+"/formatted/com/company/assembleegameclient/util/AssetLoader.as", assetLoader, 0666)

	err = exec.Command(java, "-jar", "ffdec.jar", "-selectclass", "kabam.rotmg.assets.EmbeddedAssets", "-export", "script", workingPath+"decompiled"+version, workingPath+"client"+version+".swf").Run()
	logErr(err)

	embeddedAssets, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/scripts/kabam/rotmg/assets/EmbeddedAssets.as")
	logErr(err)

	ioutil.WriteFile(workingPath+"decompiled"+version+"/formatted/kabam/rotmg/assets/EmbeddedAssets.as", embeddedAssets, 0666)

	err = exec.Command(java, "-jar", "ffdec.jar", "-selectclass", "kabam.rotmg.assets.EmbeddedData", "-export", "script", workingPath+"decompiled"+version, workingPath+"client"+version+".swf").Run()
	logErr(err)

	embeddedData, err := ioutil.ReadFile(workingPath + "decompiled" + version + "/scripts/kabam/rotmg/assets/EmbeddedData.as")
	logErr(err)

	ioutil.WriteFile(workingPath+"decompiled"+version+"/formatted/kabam/rotmg/assets/EmbeddedData.as", embeddedData, 0666)

	fmt.Println("Warning! Vanilla only!")
	fmt.Println("Config saved.")
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
	case 7:
		exportPackets()
		return
	case 8:
		exportAssetsConfig()
		return
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
	path, err := filepath.Abs("./")
	logErr(err)
	workingPath = path + "/"
	vers, err := ioutil.ReadFile(workingPath + "lib/version.txt")
	logErr(err)
	version = string(vers)
	checkMenu()
}
