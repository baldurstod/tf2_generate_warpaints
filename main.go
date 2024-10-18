package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

const PROTO_DEFS = "tf_proto_def_messages.proto"

func main() {
	var inputFile string
	var outputFile string

	flag.StringVar(&inputFile, "i", "", "Input file")
	flag.StringVar(&outputFile, "o", "", "Output file")
	flag.Parse()

	if inputFile == "" {
		fmt.Println("No input file provided. Use the flag -i")
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Println("No output file provided. Use the flag -o")
		os.Exit(1)
	}

	warpaints, err := extractWarpaints(inputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	j, err := json.Marshal(&warpaints)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.WriteFile(outputFile, j, 0666)
}

func extractWarpaints(inputFile string) (*map[string]map[string]any, error) {
	fileContent, err := os.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	nextOffset := int64(0)
	reader := bytes.NewReader(fileContent)

	var elementType int32
	var elementCount int32
	var elementSize int32
	protoElements := make(map[string]map[string]any)
	for {
		if nextOffset >= int64(reader.Len()) {
			break
		}
		reader.Seek(nextOffset, io.SeekStart)
		binary.Read(reader, binary.LittleEndian, &elementType)
		binary.Read(reader, binary.LittleEndian, &elementCount)

		elementTypeStr := strconv.Itoa(int(elementType))

		protoElements[elementTypeStr] = make(map[string]interface{})

		nextOffset, _ = reader.Seek(0, io.SeekCurrent)
		for i := int32(0); i < elementCount; i++ {
			reader.Seek(nextOffset, io.SeekStart)
			binary.Read(reader, binary.LittleEndian, &elementSize)
			elementOffset, _ := reader.Seek(0, io.SeekCurrent)
			nextOffset = elementOffset + int64(elementSize)

			buf := make([]byte, elementSize)
			_, err := reader.Read(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to read data: %w", err)
			}

			type GetHeaderProtoReflectInterface interface {
				GetHeader() *CMsgProtoDefHeader
				ProtoReflect() protoreflect.Message
			}

			structs := map[int32]func() GetHeaderProtoReflectInterface{
				7:  func() GetHeaderProtoReflectInterface { return &CMsgPaintKit_Operation{} },
				8:  func() GetHeaderProtoReflectInterface { return &CMsgPaintKit_ItemDefinition{} },
				9:  func() GetHeaderProtoReflectInterface { return &CMsgPaintKit_Definition{} },
				10: func() GetHeaderProtoReflectInterface { return &CMsgHeaderOnly{} },
			}

			if f, ok := structs[elementType]; ok {
				def := f()
				err = proto.Unmarshal(buf, def)
				if err != nil {
					return nil, fmt.Errorf("failed to Unmarshal data: %w", err)
				}
				protoElements[elementTypeStr][strconv.Itoa(int(def.GetHeader().GetDefindex()))] = def
			}
		}
	}

	return &protoElements, nil
}
