package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/chbmuc/lirc"
	"github.com/lucasb-eyer/go-colorful"
	"os"
	"os/signal"
	"log"
	"fmt"
)

func Round(val float64) int {
	if val < 0 {
		return int(val-0.5)
	}
	return int(val+0.5)
}

var colors = map[string]colorful.Color{
	"R0": colorful.Hsv(0.995,1,1), //colorful.Hsv(0.995,1,0.979),
	"G0": colorful.Hsv(0.387,0.844,1), //colorful.Hsv(0.387,0.844,0.588),
	"B0": colorful.Hsv(0.588,0.94,1), //colorful.Hsv(0.588,0.94,0.705),
	"R1": colorful.Hsv(0.06,0.942,1), //colorful.Hsv(0.06,0.942,0.857),
	"G1": colorful.Hsv(0.253,0.689,1), //colorful.Hsv(0.253,0.689,0.79),
	"B1": colorful.Hsv(0.594,0.833,1), //colorful.Hsv(0.594,0.833,0.977),
	"R2": colorful.Hsv(0.993,0.785,1), //colorful.Hsv(0.993,0.785,0.987),
	"G2": colorful.Hsv(0.49,0.871,1), //colorful.Hsv(0.49,0.871,0.541),
	"B2": colorful.Hsv(0.757,0.826,1), //colorful.Hsv(0.757,0.826,0.571),
	"R3": colorful.Hsv(0.088,0.963,1), //colorful.Hsv(0.088,0.963,0.974),
	"G3": colorful.Hsv(0.572,0.926,1), //colorful.Hsv(0.572,0.926,0.527),
	"B3": colorful.Hsv(0.72,1,1), //colorful.Hsv(0.72,1,0.998),
	"R4": colorful.Hsv(0.167,0.956,1), //colorful.Hsv(0.167,0.956,1),
	"G4": colorful.Hsv(0.598,0.935,1), //colorful.Hsv(0.598,0.935,0.33),
	"B4": colorful.Hsv(0.753,1,1), //colorful.Hsv(0.753,1,0.998),
}


type Light interface {
	toggle(on bool) error
	setHue(value float64) error
	setSaturation(value float64) error
	setBrightness(value int) error
}

type DebugLight struct {
	brightness int
	hue float64
	saturation float64
}

func MakeDebugLight() *DebugLight {
	return &DebugLight{brightness: 10}
}

func (c *DebugLight) toggle(on bool) error {
	if on == true {
		log.Println("Light on")
	} else {
		log.Println("Light off")
	}

	return nil
}

func (light *DebugLight) setBrightness(value int) error {
	brightness := Round(float64(value) / 10.0)
	direction := 1

	if brightness < light.brightness {
		direction = -1
	}

	for brightness != light.brightness {
		light.brightness += direction
		fmt.Printf("Brightness to %d\n", light.brightness)
	}

	return nil
}

func (light *DebugLight) setSaturation(value float64) error {
	light.saturation = value
	return light.setColor()
}

func (light *DebugLight) setHue(value float64) error {
	light.hue = value
	return light.setColor()
}

func (light *DebugLight) setColor() error {
	target := colorful.Hsv(light.hue / 360.0, light.saturation / 100.0, 1.0)

	var closest string

	for key, color := range colors {
		if closest == "" || colors[closest].DistanceLab(target) > color.DistanceLab(target) {
			closest = key
		}
	}

	fmt.Printf("closest color to %v: %s (%v) %f\n", target, closest, colors[closest], colors[closest].DistanceLab(target))

	return nil;
}





type IRLight struct {
	ir *lirc.Router
	brightness int
	hue float64
	saturation float64
	color string
}

func MakeIRLight(ir *lirc.Router) *IRLight {
	return &IRLight {ir: ir, brightness: 10}
}

func (light *IRLight) toggle(on bool) error {
	if on == true {
		return light.ir.Send("LED-SF-00297 POWER_ON")
	} else {
		return light.ir.Send("LED-SF-00297 POWER_OFF")
	}
}

func (light *IRLight) setBrightness(value int) error {
	brightness := Round(float64(value) / 10.0)
	direction := 1

	if brightness < light.brightness {
		direction = -1
	}

	var err error

	for err == nil && brightness != light.brightness {
		light.brightness += direction
		if direction == 1 {
			err = light.ir.Send("LED-SF-00297 BRIGHTNESS_UP")
		} else {
			err = light.ir.Send("LED-SF-00297 BRIGHTNESS_DOWN")
		}
	}

	return err
}

func (light *IRLight) setSaturation(value float64) error {
	light.saturation = value
	return light.setColor()
}

func (light *IRLight) setHue(value float64) error {
	light.hue = value
	return light.setColor()
}

func (light *IRLight) setColor() error {
	target := colorful.Hsv(light.hue / 360.0, light.saturation / 100.0, 1.0)

	var closest string

	for key, color := range colors {
		if closest == "" || colors[closest].DistanceLab(target) > color.DistanceLab(target) {
			closest = key
		}
	}

	if light.color != closest {
		light.color = closest
		return light.ir.Send(fmt.Sprintf("LED-SF-00297 %s", light.color))
	} else {
		return nil
	}
}


func main() {
	var light Light

	ir, err := lirc.Init("/var/run/lirc/lircd")

	if err == nil {
		light = MakeIRLight(ir)
	} else {
		light = MakeDebugLight()
	}

	info := accessory.Info{
		Name: "LED-SF-00279 Light strip",
		Manufacturer: "Jelmer",
	}

	acc := accessory.NewLightbulb(info)

	acc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		err := light.toggle(on)

		if err != nil {
			log.Println(err)
		}
	})

	acc.Lightbulb.Brightness.OnValueRemoteUpdate(func(value int) {
		err := light.setBrightness(value)

		if err != nil {
			log.Println(err)
		}
	})

	acc.Lightbulb.Hue.OnValueRemoteUpdate(func(value float64) {
		err := light.setHue(value)

		if err != nil {
			log.Println(err)
		}
	})

	acc.Lightbulb.Saturation.OnValueRemoteUpdate(func(value float64) {
		err := light.setSaturation(value)

		if err != nil {
			log.Println(err)
		}
	})

	t, err := hc.NewIPTransport(hc.Config{Pin: "11223344"}, acc.Accessory)

	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)

	go func() {
		<- c
		t.Stop()
		os.Exit(1)
	}()

	t.Start()
}
