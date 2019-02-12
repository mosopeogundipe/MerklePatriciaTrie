package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strings"
)

type Flag_value struct {
	encoded_prefix []uint8
	value string
}

type Node struct {
	node_type int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value Flag_value
}

type MerklePatriciaTrie struct {
	db map[string]Node
	root string
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	// TODO
	return "", errors.New("path_not_found")
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	// TODO
	var strToAscii []uint8
	var decoded []uint8
	//var encoded []uint8

	strToAscii = str_to_ascii(key)
	decoded = compact_decode_wt_prefix(strToAscii)

	//It's a leaf, so insert 16 at the end -- doing it inside insert_helper
	//decoded = append(decoded, 16)
	//encoded = compact_encode(decoded)

	root_node := mpt.insert_Helper(decoded, new_value, mpt.db[mpt.root])
	mpt.root = root_node.hash_node()
	if(mpt.db == nil){
		mpt.db = make(map[string]Node)
	}
	mpt.db[mpt.root] = root_node
	//fmt.Println("db: ",mpt.db)
	////case 1: empty db, create leaf
	//if(mpt.root == "" || len(mpt.db) == 0){
	//	strToAscii = str_to_ascii(key)
	//	decoded = compact_decode_wt_prefix(strToAscii)
	//	//It's a leaf, so insert 16 at the end
	//	decoded = append(decoded, 16)
	//	encoded = compact_encode(decoded)
	//	node :=  Node{2, [17] string{}, Flag_value{encoded, new_value }}
	//	mpt.root = node.hash_node()
	//	mpt.db = make(map[string]Node)
	//	mpt.db[mpt.root] = node
	//	fmt.Println(mpt.db)
	//}
}

func (mpt *MerklePatriciaTrie) insert_Helper(path []uint8, new_value string, current_node Node) Node{
	previous_node := current_node
	if current_node.node_type == 0  || mpt.root == "" {
		//create leaf node
		return Node{2, [17] string{}, Flag_value{compact_encode(append(path, 16)), new_value }}
		//return n
	} else if current_node.node_type == 1{
		if len(path) == 0{
			current_node.branch_value[16] = new_value
		} else{
			n := mpt.insert_Helper(path[1:], new_value, current_node)
			current_node.branch_value[path[0]] = n.hash_node()
		}
	}else if current_node.node_type == 2{
		existing_node_path := compact_decode(current_node.flag_value.encoded_prefix)
		length := len(existing_node_path)
		var index int = 0
		if len(existing_node_path) > len(path){
			length = len(path)
		}
		for i:=0; i<length; i++{
			index++
			if(existing_node_path[i] != path[i]){
				break;
			}
		}
		if index == len(existing_node_path) && index == len(path){
			if(current_node.flag_value.encoded_prefix[0] == 2 || current_node.flag_value.encoded_prefix[0] == 3){ //if current node is leaf
				current_node.flag_value.value = new_value;
			} else{ //if nodes (current node and node_to_be_inserted) have same path and current is an extension node, then just store the new value in next branch node, and  override previous
				n := mpt.insert_Helper(path[index:],new_value,mpt.db[current_node.flag_value.value])
				current_node.flag_value.value = n.hash_node()
			}
		} else if index < len(existing_node_path) && index < len(path){
			branch_node := Node{1, [17] string{""}, Flag_value{ }}
			leaf1 := Node{2, [17] string{}, Flag_value{compact_encode(append(path[index:], 16)), new_value }}
			//var node2 Node
			if(compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3){
				leaf2 := Node{2, [17] string{}, Flag_value{compact_encode(append(existing_node_path[index:], 16)), new_value }}
				mpt.db[leaf2.hash_node()] = leaf2
			} else{//change values in extension node to fit new values
				current_node.flag_value.encoded_prefix = compact_encode(existing_node_path[:index])
				current_node.flag_value.value = branch_node.hash_node()
			}
			mpt.db[branch_node.hash_node()] = branch_node
			mpt.db[leaf1.hash_node()] = leaf1
		}else if index < len(existing_node_path) && index == len(path){
			var new_node Node
			if(compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3) { //if current node is leaf
				fmt.Println("is leaf")
				new_node = Node{2, [17] string{}, Flag_value{compact_encode(append(existing_node_path[index + 1:], 16)), current_node.flag_value.value }} // make leaf of remaining path that would not be stored in extension node
			}else{
				fmt.Println("is extension")
				fmt.Println("curr node value",current_node.flag_value.value)
				new_node = Node{2, [17] string{}, Flag_value{compact_encode(append(existing_node_path[index + 1:])), current_node.flag_value.value }}
				//current_node.flag_value.value = n.hash_node()
			}
			branch_node := Node{1, [17] string{""}, Flag_value{ }}
			//check if it's leaf and extension node and create based on that
			fmt.Println("Index: ", index)
			branch_node.branch_value[16] = new_value
			branch_node.branch_value[index] = new_node.hash_node()
			fmt.Println("leaf node val: ", new_node.flag_value.value)
			current_node = Node{2, [17] string{}, Flag_value{compact_encode(path), branch_node.hash_node() }}
			fmt.Println("leaf hash: ", new_node.hash_node())
			mpt.db[new_node.hash_node()] = new_node
			mpt.db[branch_node.hash_node()] = branch_node
			return current_node
			//mpt.db[current_node.hash_node()] = current_node
		}else if index == len(existing_node_path) && index < len(path){
			var new_node Node
			if(compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3) { //if current node is leaf
				fmt.Println("is leaf")
				new_node = Node{2, [17] string{}, Flag_value{compact_encode(append(path[index + 1:], 16)), new_value }} // make leaf of remaining path that would not be stored in extension node
			}else{
				fmt.Println("is extension")
				new_node = Node{2, [17] string{}, Flag_value{compact_encode(append(path[index + 1:])), new_value }}
			}
			branch_node := Node{1, [17] string{""}, Flag_value{ }}
			//check if it's leaf and extension node and create based on that
			fmt.Println("Index: ", index)
			//leaf_node := Node{2, [17] string{}, Flag_value{compact_encode(append(path[index + 1:], 16)), new_value }} // make leaf of remaining path that would not be stored in extension node
			branch_node.branch_value[16] = current_node.flag_value.value
			branch_node.branch_value[index] = new_node.hash_node()
			current_node = Node{2, [17] string{}, Flag_value{compact_encode(existing_node_path), branch_node.hash_node() }}
			mpt.db[new_node.hash_node()] = new_node
			mpt.db[branch_node.hash_node()] = branch_node
			//sreturn current_node
			//mpt.db[current_node.hash_node()] = current_node
		}
	}
	if previous_node.hash_node() != current_node.hash_node() {
		delete(mpt.db, previous_node.hash_node())
		mpt.db[current_node.hash_node()] = current_node
	}
	return current_node
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	// TODO
	return "", errors.New("path_not_found")
}

func compact_encode(hex_array []uint8) []uint8 {
	// TODO
	var term int
	if hex_array[len(hex_array) -1] == 16 {
		term = 1
	} else{
		term = 0
	}
	if term == 1{
		hex_array = hex_array[:len(hex_array) -1]
	}
	var oddlen int = len(hex_array) % 2
	var flags []uint8 = []uint8{uint8(2 * term + oddlen)}
	if oddlen > 0{
		hex_array = append(flags, hex_array...)
	} else{
		var zeroArr []uint8 = []uint8{0}
		flags = append(flags, zeroArr...)
		hex_array = append(flags, hex_array...)
	}
	var result []uint8
	fmt.Println(hex_array)
	for i:=0; i<len(hex_array); i+=2{
		result = append(result, 16 * hex_array[i] + hex_array[i+1])
	}
	fmt.Print(term, oddlen, flags)
	return result
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	// TODO
	result := compact_decode_wt_prefix(encoded_arr)
	//fmt.Println("result", result)
	//removed prefix from compact encode below
	firstNibble := result[0]
	if(firstNibble == 0 || firstNibble ==2){
		result = result[2:]
	} else {
		result = result[1:]
	}
	return result
}

func compact_decode_wt_prefix(encoded_arr []uint8) []uint8 {
	var result []uint8
	for i:=0; i<len(encoded_arr); i+=1{
		result = append(result, encoded_arr[i]/16)
		result = append(result,  encoded_arr[i]%16)
	}
	return result
}

//takes input of ascii and gives the prefix
func get_decode_prefix(ascii_arr[] uint8) []uint8{
	result := compact_decode_wt_prefix(ascii_arr)
	firstNibble := result[0]
	if(firstNibble == 0 || firstNibble ==2){
		//fmt.Println("Even first nibble")
		result = result[0:2]
	} else {
		//fmt.Println("Odd first nibble")
		result = result[0:1]
	}
	return result
}

func str_to_ascii(input string) []uint8{
	if len(input) == 0{
		return nil
	}
	return []uint8(input)
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func main() {
	//fmt.Print("original array: ", []uint8{0, 15, 1, 12, 11, 8, 16})
	//fmt.Println(str_to_ascii("dog"))
	//val := compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})
	//val := []uint8{100,111}
	mpt := &MerklePatriciaTrie{}
	mpt.Insert("ab","10")
	mpt.Insert("c","20")
	fmt.Println("root: ",mpt.root)
	for key,value:= range mpt.db{
		fmt.Println("key: ",key, "value: ", value)
	}
	//fmt.Println(" input:", val)
	//fmt.Println("decoded array:", compact_decode([]uint8{32,100,111}))
	//fmt.Println("decoded array wt prefix:", compact_decode_wt_prefix(val))
	//fmt.Println("encoded array:", compact_encode(compact_decode_wt_prefix(val)))
	//fmt.Println("encoded array2:", compact_encode([]uint8{6,4,6,15,16}))
	//fmt.Println("prefix:", get_decode_prefix(val))
}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0] / 16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart" + cur_hash + "HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart" + cur_hash + "HashEnd", fmt.Sprintf("Hash%v", i),  -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}