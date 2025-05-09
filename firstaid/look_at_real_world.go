package firstaid

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/media"
	"github.com/use-go/onvif/ptz"
	"github.com/use-go/onvif/sdk"
	sdkmedia "github.com/use-go/onvif/sdk/media"
	sdkptz "github.com/use-go/onvif/sdk/ptz"
	xsdonvif "github.com/use-go/onvif/xsd/onvif"

	"github.com/flitsinc/go-llms/content"
	"github.com/flitsinc/go-llms/tools"
)

type LookAtRealWorldParams struct {
	RelativePan  float64 `json:"relative_pan,omitempty" description:"A value from -1.0 (right) to 1.0 (left) indicating how much to pan."`
	RelativeTilt float64 `json:"relative_tilt,omitempty" description:"A value from -1.0 (up) to 1.0 (down) indicating how much to tilt."`
	HighQuality  bool    `json:"high_quality,omitempty" description:"Use true if you want a high-resolution photo."`
}

func init() {
	sdk.Logger = sdk.Logger.Level(zerolog.Disabled)
}

var LookAtRealWorld = tools.Func(
	"Look at real world",
	"Takes a photo of the real world with the camera. Optionally pans/tilts first (camera has a 360 view). If you're looking for something, search by tilting/panning, taking a low-resolution image, and look at the result. If you don't see what you want, do another search pass, otherwise you can choose to get the high-resolution image if you want to see more.",
	"look_at_real_world",
	func(r tools.Runner, p LookAtRealWorldParams) tools.Result {
		device, err := onvif.NewDevice(onvif.DeviceParams{
			Xaddr:    os.Getenv("CAMERA_ONVIF"),
			Username: os.Getenv("CAMERA_USERNAME"),
			Password: os.Getenv("CAMERA_PASSWORD"),
		})
		if err != nil {
			return tools.ErrorWithLabel("Look at real world", fmt.Errorf("failed to connect to camera: %v", err))
		}

		profile, err := getDefaultProfile(device)
		if err != nil {
			return tools.ErrorWithLabel("Look at real world", fmt.Errorf("failed to get metadata about camera: %w", err))
		}

		if p.RelativePan != 0 || p.RelativeTilt != 0 {
			err := relativeMove(device, profile.Token, p.RelativePan, p.RelativeTilt, 0)
			if err != nil {
				return tools.ErrorWithLabel("Look at real world", fmt.Errorf("failed to pan/tilt camera: %v", err))
			}
		}

		photoPath, err := takePhoto()
		if err != nil {
			return tools.ErrorWithLabel("Look at real world", fmt.Errorf("failed to get photo path: %v", err))
		}
		defer os.Remove(photoPath)
		imageName, dataURI, err := content.ImageToDataURI(photoPath, p.HighQuality)
		if err != nil {
			return tools.ErrorWithLabel("Look at real world", fmt.Errorf("failed to process image %s: %w", imageName, err))
		}
		resultContent := content.Content{&content.ImageURL{URL: dataURI}}
		return tools.SuccessWithContent("Look at real world", resultContent)
	},
)

func getDefaultProfile(device *onvif.Device) (xsdonvif.Profile, error) {
	res, err := sdkmedia.Call_GetProfiles(context.TODO(), device, media.GetProfiles{})
	if err != nil {
		return xsdonvif.Profile{}, err
	}
	if len(res.Profiles) == 0 {
		return xsdonvif.Profile{}, fmt.Errorf("no profiles found")
	}
	return res.Profiles[0], nil
}

func relativeMove(device *onvif.Device, token xsdonvif.ReferenceToken, pan, tilt, zoom float64) error {
	req := ptz.RelativeMove{
		ProfileToken: token,
		Translation: xsdonvif.PTZVector{
			PanTilt: xsdonvif.Vector2D{
				X:     pan,
				Y:     tilt,
				Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace",
			},
			Zoom: xsdonvif.Vector1D{
				X:     zoom,
				Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/TranslationGenericSpace",
			},
		},
	}
	if _, err := sdkptz.Call_RelativeMove(context.TODO(), device, req); err != nil {
		return err
	}
	// Wait up to 10 seconds for the pan/tilt to complete.
	beganWaiting := time.Now()
	for time.Since(beganWaiting) < 10*time.Second {
		time.Sleep(100 * time.Millisecond)
		res, err := device.CallMethod(ptz.GetStatus{ProfileToken: token})
		if err != nil {
			return err
		}
		type PTZStatusResponse struct {
			PanTiltStatus string `xml:"Body>GetStatusResponse>PTZStatus>MoveStatus>PanTilt"`
			ZoomStatus    string `xml:"Body>GetStatusResponse>PTZStatus>MoveStatus>Zoom"`
		}
		var status PTZStatusResponse
		if err := xml.NewDecoder(res.Body).Decode(&status); err != nil {
			return err
		}
		if status.PanTiltStatus == "idle" {
			break
		}
	}
	return nil
}

func takePhoto() (string, error) {
	// Build the RTSP URI with username and password included.
	u, err := url.Parse(os.Getenv("CAMERA_RTSP"))
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(os.Getenv("CAMERA_USERNAME"), os.Getenv("CAMERA_PASSWORD"))
	// Create a temporary path to write the snapshot to.
	photoPath := fmt.Sprintf("%s/snapshot_%d.jpg", os.TempDir(), time.Now().Unix())
	// Use ffmepg to read one frame from the RTSP stream.
	cmd := exec.Command("ffmpeg", "-loglevel", "error", "-i", u.String(), "-f", "image2", "-vframes", "1", "-pix_fmt", "yuvj420p", photoPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return photoPath, nil
}
