package actions

import (
	"log"
	"os/exec"
)

type RFSwitchOn struct {
	rfServiceName string
}

func NewRFSwitchOn(serviceName string) *RFSwitchOn {
	return &RFSwitchOn{
		rfServiceName: serviceName,
	}
}

// HandleParamValue Обработка изменений ключевого параметра
func (rf *RFSwitchOn) HandleParamValue(value float32) {
	var action string = "unknown"

	switch value {
	case 0:

		action = "stop"
		log.Println("RF switch off")
	case 1:
		action = "start"
		log.Println("RF switch on")
	}

	if action == "unknown" {
		log.Println("RF switch unknown param value")
		return
	}

	cmd := exec.Command("sudo", "systemctl", action, rf.rfServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command failed: %s\nError: %v\n", output, err)
		return
	}
	log.Printf("Command succeeded: %s\n", output)
}
