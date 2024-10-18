# tf2_generate_warpaints

Generating tf_proto_def_messages.pb.go : protoc --go_out=./ --go_opt=Mtf_proto_def_messages.proto=github.com/baldurstod/tf2_generate_warpaints/main tf_proto_def_messages.proto


Usage: tf2_generate_warpaints -i ./var/proto_defs.vpd -p ./var/tf_proto_def_messages.proto -o ./var/warpaints.json
