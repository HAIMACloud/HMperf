//modify from https://github.com/shogo82148/androidbinary/apk
//original author: shogo82148
//support android8.0+ adaptive icon: background & foreground image merge to 1 jpeg picture

package apk

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/shogo82148/androidbinary"

	_ "image/jpeg" // handle jpeg format
	_ "image/png"  // handle png format

	_ "golang.org/x/image/webp"
)

// Apk is an application package file for android.
type Apk struct {
	f         *os.File
	zipreader *zip.Reader
	manifest  Manifest
	table     *androidbinary.TableFile
}

// OpenFile will open the file specified by filename and return Apk
func OpenFile(filename string) (apk *Apk, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	apk, err = OpenZipReader(f, fi.Size())
	if err != nil {
		return nil, err
	}
	apk.f = f
	return
}

// OpenZipReader has same arguments like zip.NewReader
func OpenZipReader(r io.ReaderAt, size int64) (*Apk, error) {
	zipreader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	apk := &Apk{
		zipreader: zipreader,
	}
	if err = apk.parseResources(); err != nil {
		return nil, err
	}
	if err = apk.parseManifest(); err != nil {
		return nil, errorf("parse-manifest: %w", err)
	}
	return apk, nil
}

// Close is avaliable only if apk is created with OpenFile
func (k *Apk) Close() error {
	if k.f == nil {
		return nil
	}
	return k.f.Close()
}

func (k *Apk) IconPath(resConfig *androidbinary.ResTableConfig) ([]string, error) {
	iconPath, err := k.manifest.App.Icon.WithResTableConfig(resConfig).String()
	if err != nil {
		return nil, err
	}
	if androidbinary.IsResID(iconPath) {
		return nil, newError("unable to convert icon-id to icon path")
	}
	suffix := filepath.Ext(iconPath)
	if suffix == ".xml" {
		bz, _ := k.ReadZipFile(iconPath)
		xmlFile, err := androidbinary.NewXMLFile(bytes.NewReader(bz))
		if err != nil {
			return nil, err
		}
		adaptIconData := new(AdaptiveIcon)
		if err := xmlFile.Decode(adaptIconData, k.table, nil); err != nil {
			return nil, err
		}
		iconPathBackground, _ := adaptIconData.Background.Drawable.String()
		iconPathForeground, _ := adaptIconData.Foreground.Drawable.String()
		return []string{iconPathBackground, iconPathForeground}, nil
	}
	return []string{iconPath}, nil
}

func (k *Apk) getImage(iconPath string) (image.Image, error) {
	imgData, err := k.ReadZipFile(iconPath)
	if err != nil {
		return nil, err
	}
	m, _, err := image.Decode(bytes.NewReader(imgData))
	return m, err
}

// IconJpeg Icon returns the icon image of the APK.
func (k *Apk) IconJpeg(resConfig *androidbinary.ResTableConfig) ([]byte, error) {
	iconPath, err := k.IconPath(resConfig)
	if err != nil {
		return nil, err
	}
	var output = new(bytes.Buffer)
	var byteWriter = bufio.NewWriter(output)
	var tmp *image.RGBA
	if len(iconPath) == 1 {
		img, err := k.getImage(iconPath[0])
		if err != nil {
			return nil, err
		}
		tmp = image.NewRGBA(img.Bounds())
		draw.Draw(tmp, img.Bounds(), img, img.Bounds().Min, draw.Src)
	} else if len(iconPath) == 2 {
		backgroundImg, err1 := k.getImage(iconPath[0])
		foregroundImg, err2 := k.getImage(iconPath[1])
		if err1 == nil && err2 != nil {
			tmp = image.NewRGBA(backgroundImg.Bounds())
			draw.Draw(tmp, backgroundImg.Bounds(), backgroundImg, backgroundImg.Bounds().Min, draw.Over)
		} else if err1 != nil && err2 == nil {
			tmp = image.NewRGBA(foregroundImg.Bounds())
			draw.Draw(tmp, foregroundImg.Bounds(), foregroundImg, foregroundImg.Bounds().Min, draw.Over)
		} else if err1 == nil && err2 == nil {
			tmp = image.NewRGBA(backgroundImg.Bounds())
			draw.Draw(tmp, backgroundImg.Bounds(), backgroundImg, backgroundImg.Bounds().Min, draw.Over)
			draw.Draw(tmp, foregroundImg.Bounds(), foregroundImg, foregroundImg.Bounds().Min, draw.Over)
		} else if err1 != nil && err2 != nil {
			return nil, fmt.Errorf("err1:%s, err2:%s", err1.Error(), err2.Error())
		}
	}
	if err := jpeg.Encode(byteWriter, tmp, nil); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// Label returns the label of the APK.
func (k *Apk) Label(resConfig *androidbinary.ResTableConfig) (s string, err error) {
	s, err = k.manifest.App.Label.WithResTableConfig(resConfig).String()
	if err != nil {
		return
	}
	if androidbinary.IsResID(s) {
		err = newError("unable to convert label-id to string")
	}
	return
}

// Manifest returns the manifest of the APK.
func (k *Apk) Manifest() Manifest {
	return k.manifest
}

// PackageName returns the package name of the APK.
func (k *Apk) PackageName() string {
	return k.manifest.Package.MustString()
}

func isMainIntentFilter(intent ActivityIntentFilter) bool {
	ok := false
	for _, action := range intent.Actions {
		s, err := action.Name.String()
		if err == nil && s == "android.intent.action.MAIN" {
			ok = true
			break
		}
	}
	if !ok {
		return false
	}
	ok = false
	for _, category := range intent.Categories {
		s, err := category.Name.String()
		if err == nil && s == "android.intent.category.LAUNCHER" {
			ok = true
			break
		}
	}
	return ok
}

// MainActivity returns the name of the main activity.
func (k *Apk) MainActivity() (activity string, err error) {
	for _, act := range k.manifest.App.Activities {
		for _, intent := range act.IntentFilters {
			if isMainIntentFilter(intent) {
				return act.Name.String()
			}
		}
	}
	for _, act := range k.manifest.App.ActivityAliases {
		for _, intent := range act.IntentFilters {
			if isMainIntentFilter(intent) {
				return act.TargetActivity.String()
			}
		}
	}

	return "", newError("No main activity found")
}

func (k *Apk) parseManifest() error {
	xmlData, err := k.ReadZipFile("AndroidManifest.xml")
	if err != nil {
		return errorf("failed to read AndroidManifest.xml: %w", err)
	}
	xmlfile, err := androidbinary.NewXMLFile(bytes.NewReader(xmlData))
	if err != nil {
		return errorf("failed to parse AndroidManifest.xml: %w", err)
	}
	return xmlfile.Decode(&k.manifest, k.table, nil)
}

func (k *Apk) parseResources() (err error) {
	resData, err := k.ReadZipFile("resources.arsc")
	if err != nil {
		return
	}
	k.table, err = androidbinary.NewTableFile(bytes.NewReader(resData))
	return
}

func (k *Apk) ReadZipFile(name string) (data []byte, err error) {
	buf := bytes.NewBuffer(nil)
	for _, file := range k.zipreader.File {
		if file.Name != name {
			continue
		}
		rc, er := file.Open()
		if er != nil {
			err = er
			return
		}
		defer rc.Close()

		_, err = io.Copy(buf, rc)
		if err != nil {
			return
		}
		return buf.Bytes(), nil
	}
	return nil, fmt.Errorf("file %s not found", strconv.Quote(name))
}
