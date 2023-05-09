package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

func main() {
	set, err := processDescriptorDir([]string{"./proto"})
	if err != nil {
		panic(err)
	}
	desc, err := GetAllMessageDescriptorsInFDS(set)
	if err != nil {
		panic(err)
	}
	fac := dynamic.NewMessageFactoryWithDefaults()
	resolver := dynamic.AnyResolver(fac, desc...)
	unmarshaler := &jsonpb.Unmarshaler{AnyResolver: resolver}
	envelopMD, err := FindMessageDescriptorInFDS(set, "")
	if err != nil {
		panic(err)
	}
	envelope := fac.NewDynamicMessage(envelopMD)
	envelope.UnmarshalJSONPB(unmarshaler, []byte("your json content"))
	protoBytes, err := envelope.Marshal()
	if err != nil {
		panic(err)
	}
	fmt.Println(protoBytes)
}

func GetAllMessageDescriptorsInFDS(fds *dpb.FileDescriptorSet) ([]*desc.FileDescriptor, error) {
	allDescriptors := make([]*desc.FileDescriptor, 0)

	descriptors, err := desc.CreateFileDescriptorsFromSet(fds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file descriptor set")
	}

	for _, d := range descriptors {
		allDescriptors = append(allDescriptors, d)
	}

	return allDescriptors, nil
}
func processDescriptorDir(dirs []string) (*dpb.FileDescriptorSet, error) {
	files, err := getProtoFiles(dirs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get proto files")
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no .proto found in dir(s) '%v'", dirs)
	}

	fds, err := readFileDescriptorsV1(files)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read file descriptors")
	}

	descriptorSet := desc.ToFileDescriptorSet(fds...)

	return descriptorSet, nil
}
func getProtoFiles(dirs []string) (map[string][]string, error) {
	protos := make(map[string][]string, 0)

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("unable to walk path '%s': %s", dir, err)
			}

			if info.IsDir() {
				// Nothing to do if this is a dir
				return nil
			}

			if strings.HasSuffix(info.Name(), ".proto") {
				if _, ok := protos[dir]; !ok {
					protos[dir] = make([]string, 0)
				}

				protos[dir] = append(protos[dir], path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking the path '%s': %v", dir, err)
		}
	}

	return protos, nil
}
func readFileDescriptorsV1(files map[string][]string) ([]*desc.FileDescriptor, error) {
	contents := make(map[string]string, 0)
	keys := make([]string, 0)

	for dir, files := range files {
		// cleanup dir
		dir = filepath.Clean(dir)

		for _, f := range files {
			data, err := ioutil.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("unable to read file '%s': %s", f, err)
			}

			if !strings.HasSuffix(dir, "/") {
				dir = dir + "/"
			}

			// Strip base path
			relative := strings.Split(f, dir)

			if len(relative) != 2 {
				return nil, fmt.Errorf("unexpected length of split path (%d)", len(relative))
			}

			contents[relative[1]] = string(data)
			keys = append(keys, relative[1])
		}
	}

	var p protoparse.Parser

	p.Accessor = protoparse.FileContentsFromMap(contents)

	fds, err := p.ParseFiles(keys...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse files")
	}

	return fds, nil
}

func FindMessageDescriptorInFDS(fds *dpb.FileDescriptorSet, messageName string) (*desc.MessageDescriptor, error) {
	descriptors, err := desc.CreateFileDescriptorsFromSet(fds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file descriptor set")
	}

	for _, d := range descriptors {
		types := d.GetMessageTypes()
		for _, md := range types {
			if md.GetFullyQualifiedName() == messageName {
				return md, nil
			}
		}
	}

	return nil, errors.New("message descriptor not found in file descriptor(s)")
}
