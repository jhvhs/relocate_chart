package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"regexp"
)

func relocateChart(chartPath string) error {
	srcChart, err := os.Open(chartPath)
	if err != nil {
		return err
	}
	defer srcChart.Close()

	srcGz, err := gzip.NewReader(srcChart)
	if err != nil {
		return err
	}
	defer srcGz.Close()
	srcTar := tar.NewReader(srcGz)

	dstChart, err := os.Create(fmt.Sprintf("%s.tgz", chartPath))
	if err != nil {
		return err
	}
	defer dstChart.Close()

	dstGz := gzip.NewWriter(dstChart)
	defer dstGz.Close()

	dstTar := tar.NewWriter(dstGz)
	defer dstTar.Close()

	for {
		err = processNextFile(srcTar, dstTar)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}
	return nil
}

func processNextFile(srcTar *tar.Reader, dstTar *tar.Writer) error {
	hdr, err := srcTar.Next()
	if err != nil {
		return err
	}
	fileName := hdr.Name

	var sourceFileReader io.Reader = srcTar

	if isMainValuesFile(fileName) {
		sourceFileReader, err = updatedValuesFileReader(srcTar, hdr)
		if err != nil {
			return err
		}
	}

	if err = dstTar.WriteHeader(hdr); err != nil {
		return err
	}

	if _, err = io.Copy(dstTar, sourceFileReader); err != nil {
		return err
	}
	return nil
}

func updatedValuesFileReader(srcTar *tar.Reader, hdr *tar.Header) (io.Reader, error) {
	values, err := ioutil.ReadAll(srcTar)
	if err != nil {
		return nil, err
	}

	updatedValues, err := updatedValuesContents(values)
	if err != nil {
		return nil, err
	}
	newValues := bytes.NewReader(updatedValues)
	hdr.Size = int64(len(updatedValues))
	return newValues, nil
}

func updatedValuesContents(values []byte) ([]byte, error) {
	var configValues map[string]interface{}
	if err := yaml.Unmarshal(values, &configValues); err != nil {
		return nil, err
	}

	var globalSection map[interface{}]interface{}
	if g, ok := configValues["global"]; ok {
		globalSection = g.(map[interface{}]interface{})
		globalSection["imageRegistry"] = newRegistry
		globalSection["imageNamespace"] = newNamespace
	} else {
		globalSection = map[interface{}]interface{}{
			"imageRegistry":  newRegistry,
			"imageNamespace": newNamespace,
		}
	}
	configValues["global"] = globalSection
	updatedValues, err := yaml.Marshal(configValues)
	if err != nil {
		return nil, err
	}

	return updatedValues, nil
}

func isMainValuesFile(fileName string) bool {
	valuesLocator, _ := regexp.Compile(`^[^\\/]+?[\\/]values.yaml$`)
	isMainValuesFile := valuesLocator.MatchString(fileName)
	return isMainValuesFile
}

