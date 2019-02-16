package p1

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
	value          string
}

type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	db   map[string]Node
	root string
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	var strToAscii []uint8
	var decoded []uint8
	strToAscii = str_to_ascii(key)
	decoded = compact_decode_wt_prefix(strToAscii)
	root := mpt.db[mpt.root]
	value := mpt.get_helper(root, decoded)
	return value, errors.New("path_not_found")
}

func (mpt MerklePatriciaTrie) get_helper(current_node Node, new_path []uint8) string {
	if current_node.node_type == 0 {
		return "failure"
	} else if current_node.node_type == 1 {
		if len(new_path) == 0 {
			return current_node.branch_value[16]
		} else {
			return mpt.get_helper(mpt.db[current_node.branch_value[new_path[0]]], new_path[1:])
		}
	} else {
		var i int
		old_path := compact_decode(current_node.flag_value.encoded_prefix)
		for i = 0; i < len(new_path) && i < len(old_path); i += 1 {
			if new_path[i] != old_path[i] {
				break
			}
		}
		if i == len(new_path) && i == len(old_path) {
			return current_node.flag_value.value
		} else if i < len(old_path) && i < len(new_path) {
			return "failure"
		} else {
			return mpt.get_helper(mpt.db[current_node.flag_value.value], new_path[i:])
		}
	}
}

//func (mpt *MerklePatriciaTrie) get_helper(path []uint8, node Node) string{

//}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	var strToAscii []uint8
	var decoded []uint8
	//var encoded []uint8

	strToAscii = str_to_ascii(key)
	decoded = compact_decode_wt_prefix(strToAscii)
	//It's a leaf, so insert 16 at the end -- doing it inside insert_helper
	//decoded = append(decoded, 16)
	//encoded = compact_encode(decoded)

	root_node := mpt.insert_helper(decoded, new_value, mpt.db[mpt.root])
	fmt.Println("root hash: ", root_node.hash_node())
	fmt.Println("root encoded: ", root_node.flag_value.encoded_prefix)
	mpt.root = root_node.hash_node()
	if mpt.db == nil {
		mpt.db = make(map[string]Node)
	}
	mpt.db[mpt.root] = root_node

	fmt.Println("root: ", mpt.root)
	for key, value := range mpt.db {
		fmt.Println("key: ", key, "value: ", value)
	}
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

func (mpt *MerklePatriciaTrie) is_any_match(decoded_path_1 []uint8, decoded_path_2 []uint8) (bool, int) {
	length := len(decoded_path_1)
	var index int = 0
	if len(decoded_path_1) > len(decoded_path_2) {
		length = len(decoded_path_2)
	}
	for i := 0; i < length; i++ {
		//index++
		if decoded_path_1[i] != decoded_path_2[i] {
			break
		}
		index++
	}
	if index > 0 {
		return true, index
	}
	return false, index
}

func (mpt *MerklePatriciaTrie) insert_helper(path []uint8, new_value string, current_node Node) Node {
	previous_node := current_node
	fmt.Println("new_ path: ", path, "value: ", new_value)
	if current_node.node_type == 0 || mpt.root == "" {
		//create leaf node
		fmt.Println("creating leaf condition 1")
		newnode := Node{2, [17]string{}, Flag_value{compact_encode(append(path, 16)), new_value}}
		current_node = newnode
		//return n
	} else if current_node.node_type == 1 {
		fmt.Println("Current is Branch")
		if len(path) == 0 {
			current_node.branch_value[16] = new_value
		} else {
			if current_node.branch_value[path[0]] != "" {
				fmt.Println("GOT IN...")
				n := mpt.insert_helper(path[1:], new_value, mpt.db[current_node.branch_value[path[0]]])
				current_node.branch_value[path[0]] = n.hash_node()
				fmt.Println("finished recursing branch, hashnode gotten: ", n.hash_node())
				fmt.Println("finished recursing branch, node gotten: ", n)
				//return current_node //HINT.TEST:
			}
		}
	} else if current_node.node_type == 2 {
		existing_node_path := compact_decode(current_node.flag_value.encoded_prefix)
		fmt.Println("existing_ path:", existing_node_path)
		length := len(existing_node_path)
		var index int = 0
		if len(existing_node_path) > len(path) {
			length = len(path)
		}
		for i := 0; i < length; i++ {
			//index++
			if existing_node_path[i] != path[i] {
				break
			}
			index++
		}
		fmt.Println("index", index, "path length: ", len(path))
		if index == len(existing_node_path) && index == len(path) { //full match in both paths (hex values)
			fmt.Println("check 0")
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				current_node.flag_value.value = new_value
			} else { //if nodes (current node and node_to_be_inserted) have same path and current is an extension node, then just store the new value in next branch node, and  override previous
				n := mpt.insert_helper(path[index:], new_value, mpt.db[current_node.flag_value.value])
				current_node.flag_value.value = n.hash_node()
			}
		} else if index == len(existing_node_path) && index < len(path) {
			fmt.Println("check 1")
			new_node := Node{2, [17]string{}, Flag_value{compact_encode(append(path[index+1:], 16)), new_value}} // make leaf of remaining path that would not be stored in extension node
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				fmt.Println("is leaf")
				if len(existing_node_path) == 0 {
					fmt.Println("path length is zero")
					branch_node.branch_value[16] = current_node.flag_value.value
					branch_node.branch_value[path[index]] = new_node.hash_node() //HINT.TEST still testing
					mpt.db[branch_node.hash_node()] = branch_node
					mpt.db[new_node.hash_node()] = new_node
					delete(mpt.db, current_node.hash_node())
					return branch_node
				} else {
					fmt.Println("path length more than zero")
					current_node.flag_value.encoded_prefix = compact_encode(existing_node_path[:index])
					branch_node.branch_value[16] = current_node.flag_value.value
					current_node.flag_value.value = branch_node.hash_node()
					mpt.db[current_node.hash_node()] = current_node
				}
			} else {
				fmt.Println("is extension")
				new_node = mpt.insert_helper(path[index:], new_value, mpt.db[current_node.flag_value.value])
				fmt.Println("NEW NODE:", new_node)
				//current_node = new_node
				current_node.flag_value.value = new_node.hash_node()
				fmt.Println("CurrentNode: ", current_node)
				fmt.Println("Current Node hash:", current_node.hash_node())
				//return current_node
			}
			fmt.Println("Index: ", index)
			branch_node.branch_value[path[index]] = new_node.hash_node()
			mpt.db[new_node.hash_node()] = new_node
			mpt.db[branch_node.hash_node()] = branch_node
			//sreturn current_node
			//mpt.db[current_node.hash_node()] = current_node
		} else if index < len(existing_node_path) && index == len(path) {
			fmt.Println("check 2")
			var new_node Node
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				fmt.Println("is leaf")
				new_node = Node{2, [17]string{}, Flag_value{compact_encode(append(existing_node_path[index+1:], 16)), current_node.flag_value.value}} // make leaf of remaining path that would not be stored in extension node
			} else {
				fmt.Println("is extension")
				fmt.Println("curr node value", current_node.flag_value.value)
				new_node = mpt.insert_helper(path[index:], new_value, mpt.db[current_node.flag_value.value])
				//current_node = new_node
				current_node.flag_value.value = new_node.hash_node()
				//current_node.flag_value.value = new_node.hash_node()
				//new_node = Node{2, [17] string{}, Flag_value{compact_encode(append(existing_node_path[index + 1:])), current_node.flag_value.value }}
			}
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			//check if it's leaf and extension node and create based on that
			fmt.Println("Index: ", index)
			branch_node.branch_value[16] = new_value
			//branch_node.branch_value[index] = new_node.hash_node()
			branch_node.branch_value[existing_node_path[index]] = new_node.hash_node()
			fmt.Println("leaf node val: ", new_node.flag_value.value)
			if len(path) > 0 {
				delete(mpt.db, current_node.hash_node()) //JUST ADDED THIS
				current_node = Node{2, [17]string{}, Flag_value{compact_encode(path), branch_node.hash_node()}}
				mpt.db[current_node.hash_node()] = current_node //JUST ADDED THIS
			}

			fmt.Println("leaf hash: ", new_node.hash_node())
			mpt.db[new_node.hash_node()] = new_node
			mpt.db[branch_node.hash_node()] = branch_node
			return branch_node
			//return current_node
			//mpt.db[current_node.hash_node()] = current_node
			//mpt.db[current_node.hash_node()] = current_node
		} else if index == 0 {
			fmt.Println("check 3")
			//new_node := Node{2, [17] string{}, Flag_value{compact_encode(append(path[index + 1:], 16)), new_value }} // make leaf of remaining path that would not be stored in extension node
			var new_node Node
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				fmt.Println("is leaf")
				new_node = Node{2, [17]string{}, Flag_value{compact_encode(append(existing_node_path[index+1:], 16)), current_node.flag_value.value}}
				mpt.db[new_node.hash_node()] = new_node
				branch_node.branch_value[existing_node_path[index]] = new_node.hash_node()
			} else { //extension
				fmt.Print("is extension")
				if len(existing_node_path) == 1 {
					branch_node.branch_value[existing_node_path[index]] = current_node.hash_node()
					delete(mpt.db, current_node.hash_node())
				} else {
					new_node = Node{2, [17]string{}, Flag_value{compact_encode(existing_node_path[index+1:]), current_node.flag_value.value}}
					mpt.db[new_node.hash_node()] = new_node
				}
			}
			leaf_node := Node{2, [17]string{}, Flag_value{compact_encode(append(path[index+1:], 16)), new_value}}
			branch_node.branch_value[path[index]] = leaf_node.hash_node()
			mpt.db[leaf_node.hash_node()] = leaf_node
			mpt.db[branch_node.hash_node()] = branch_node
			return branch_node
		} else {
			fmt.Println("check 4")
			fmt.Println("Any other case enters here")
			var new_node Node
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				fmt.Println("is leaf")
				new_node = Node{2, [17]string{}, Flag_value{compact_encode(append(existing_node_path[index+1:], 16)), current_node.flag_value.value}}
			} else { //extension
				fmt.Println("is extension")
				new_node = Node{2, [17]string{}, Flag_value{compact_encode(existing_node_path[index+1:]), current_node.flag_value.value}}
			}
			prefix := compact_encode(path[:index])
			current_node.flag_value.encoded_prefix = prefix
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			mpt.db[new_node.hash_node()] = new_node
			//branch_node.branch_value[existing_node_path[index]] = current_node.hash_node()
			branch_node.branch_value[existing_node_path[index]] = new_node.hash_node()
			leaf_node := Node{2, [17]string{}, Flag_value{compact_encode(append(path[index+1:], 16)), new_value}}
			branch_node.branch_value[path[index]] = leaf_node.hash_node()
			mpt.db[leaf_node.hash_node()] = leaf_node
			current_node.flag_value.value = branch_node.hash_node()
			mpt.db[current_node.hash_node()] = current_node
			mpt.db[branch_node.hash_node()] = branch_node
			fmt.Println("Current node: ", current_node)
			fmt.Println("Current node hash: ", current_node.hash_node())
			fmt.Println("New node: ", new_node)
			fmt.Println("New node hash: ", new_node.hash_node())
			//current_node = new_node
			return current_node //SEE IF THIS DOESN"T AFFECT OTHER CASES (works for p, aa, ap)
		}
	}
	if previous_node.hash_node() != current_node.hash_node() {
		fmt.Println("remove node :", previous_node)
		fmt.Println("hash of remove node :", previous_node.hash_node())
		delete(mpt.db, previous_node.hash_node())
		mpt.db[current_node.hash_node()] = current_node
	}
	return current_node
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	var strToAscii []uint8
	var decoded []uint8
	strToAscii = str_to_ascii(key)
	decoded = compact_decode_wt_prefix(strToAscii)
	if mpt.root == "" || len(mpt.db) == 0 {
		return "", errors.New("path_not_found")
	}
	root := mpt.db[mpt.root]
	node := mpt.delete_helper(decoded, root)
	if node.node_type == 0 {
		return "", errors.New("path_not_found")
	}
	fmt.Println("successfully deleted")
	fmt.Println("Node returned from delete_helper: ", node)
	mpt.root = node.hash_node() //?? just added, hope it's fine
	fmt.Println("CHECK root: ", mpt.root)
	for key, value := range mpt.db {
		fmt.Println("key: ", key, "value: ", value)
	}
	return "", errors.New("")
}

//find index of branch node
func find_index_branch(current_node Node) []uint8 {
	var index_of_branch []uint8
	fmt.Println(len(current_node.branch_value))
	//******HINT.SOPE : For loop on top was giving out of range exceptions
	for i := 0; i < len(current_node.branch_value); i++ {
		if current_node.branch_value[i] != "" {
			index_of_branch = append(index_of_branch, uint8(i))
		}
	}
	return index_of_branch
}

//**** HINT.SOPE: Changed the rebalance trie function get current branch indexes instead of count
func (mpt *MerklePatriciaTrie) rebalance_trie(current_node Node) Node {
	fmt.Println("GETS HERE!")
	index_of_branch := find_index_branch(current_node)
	if len(index_of_branch) == 1 {
		next_node := mpt.db[current_node.branch_value[index_of_branch[0]]]
		fmt.Println("Index of Branch:", index_of_branch[0])
		fmt.Println("next node =", current_node)
		if next_node.node_type == 2 {
			fmt.Println("next node is leaf or extension.....")
			if compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] != 1 || compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] != 0 {
				// if leaf, pull up leaf value and convert branch to leaf
				current_node.flag_value.value = next_node.flag_value.value
				fmt.Println("flg: ", next_node.flag_value.value)
				prefix := append(index_of_branch, compact_decode(next_node.flag_value.encoded_prefix)...)
				current_node.flag_value.encoded_prefix = compact_encode(append(prefix, 16))
				//current_node.flag_value.encoded_prefix = compact_encode(append(index_of_branch[0], compact_decode(next_node.flag_value.encoded_prefix)...))
				current_node.node_type = 2
				fmt.Println("CURR NODE: ", current_node)
				current_node.branch_value = [17]string{}
				//mpt.db[current_node.hash_node()] = current_node
				return current_node
			} else if compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] == 1 || compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] == 0 {
				//if extension den merge 2 extension together
				next_node_extension := compact_decode(next_node.flag_value.encoded_prefix)
				new_extension_prefix := append(compact_decode(current_node.flag_value.encoded_prefix), next_node_extension...)
				current_node.flag_value.encoded_prefix = compact_encode(new_extension_prefix) //merge 2 extension into one
				current_node.flag_value.value = next_node.flag_value.value
				return current_node
			}
		} else if next_node.node_type == 1 {
			//if next node is branch node den pull branch up and convert to extension
			index_of_new_branch := find_index_branch(next_node)
			if len(index_of_new_branch) == 1 {
				current_node.flag_value.encoded_prefix = compact_encode(append(index_of_branch, index_of_new_branch...))
				current_node.flag_value.value = current_node.branch_value[index_of_new_branch[0]]
				current_node.node_type = 2
				current_node.branch_value = [17]string{}
				return current_node
			}
		}
	}
	return current_node
}

func (mpt *MerklePatriciaTrie) delete_helper(path []uint8, current_node Node) Node {
	fmt.Println("********************************DELETING********************************")
	var previous_node Node
	previous_node = current_node
	if current_node.node_type == 1 { //branch node
		if len(path) == 0 {
			fmt.Println("entered in branch where path len == 0, node: ", current_node)
			current_node.branch_value[16] = ""
			assigned_branch_indices := find_index_branch(current_node)
			if len(assigned_branch_indices) == 1 {
				node := mpt.rebalance_trie(current_node)
				current_node = node
			} //else{//branch count > 1
			//mpt.db[current_node.hash_node()] = current_node
			//}
		} else {
			//decided how to recurse below
			fmt.Println("entered in branch where path len > 0, node: ", current_node)
			var node Node
			//if(compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 1 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2){//if extension, recurse to child
			//	node = mpt.delete_helper(path[1:], mpt.db[current_node.flag_value.value])
			//}else if(current_node.node_type == 1){//if branch, get child as zeroth index of path
			//	node = mpt.delete_helper(path[1:], mpt.db[current_node.branch_value[path[0]]])
			//}
			fmt.Println("recurse further from branch, path: ", path[1:])
			node = mpt.delete_helper(path[1:], mpt.db[current_node.branch_value[path[0]]])
			fmt.Println("entered in branch where path len > 0, result node: ", node)
			if node.node_type == 0 {
				current_node.branch_value[path[0]] = ""
				fmt.Println("Set curr node index to empty. curr node: ", current_node)
				if current_node.branch_value[16] != "" {
					fmt.Println("Entered IF in branch where path len > 0")
					var new_path []uint8
					new_node := Node{2, [17]string{}, Flag_value{compact_encode(append(new_path, 16)), current_node.branch_value[16]}} //convert to leaf node
					current_node = new_node
				} else {
					new_node := mpt.rebalance_trie(current_node)
					current_node = new_node
				}
			} else {
				fmt.Println("Entered ELSE in branch where path len > 0")
				//set current branch index = nextnode.hashnode
				current_node.branch_value[path[0]] = node.hash_node()
			}

		}
	} else if current_node.node_type == 2 {
		existing_node_path := compact_decode(current_node.flag_value.encoded_prefix)
		fmt.Println("existing_ path:", existing_node_path)
		length := len(existing_node_path)
		var index int = 0
		if len(existing_node_path) > len(path) {
			length = len(path)
		}
		for i := 0; i < length; i++ {
			if existing_node_path[i] != path[i] {
				break
			}
			index++
		}
		fmt.Println("index", index, "path length: ", len(path))
		if index == len(path) && index == len(existing_node_path) && (compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3) { //if paths match fully and node isLeaf
			fmt.Println("full match deleting")
			delete(mpt.db, current_node.hash_node()) //?????????????
			current_node.flag_value.value = ""
			current_node.flag_value.encoded_prefix = nil
			current_node.node_type = 0
			fmt.Println("after deletion := ", current_node)
		} else if index == len(existing_node_path) && (compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 1 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 0) { //is extension
			fmt.Println("partial match deleting from path: ", path[index:])
			node := mpt.delete_helper(path[index:], mpt.db[current_node.flag_value.value])
			fmt.Println("finished recursing from extension parent. returned value: ", node)
			fmt.Println("current node: ", current_node)
			fmt.Println("node: ", node)
			fmt.Println("node hash: ", node.hash_node())
			if current_node.flag_value.value != node.hash_node() {
				if node.node_type == 0 {
					fmt.Println("Entered zeroth condition")
					current_node.flag_value.value = ""
					current_node = Node{}
				} else if node.node_type == 1 {
					fmt.Println("Entered one condition")
					current_node.flag_value.value = node.hash_node()
				} else if node.node_type == 2 {
					if compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 3 { //is leaf
						fmt.Println("Entered two condition -- leaf")
						decoded_prefix1 := compact_decode(current_node.flag_value.encoded_prefix)
						decoded_prefix2 := compact_decode(node.flag_value.encoded_prefix)
						//current_node.flag_value.encoded_prefix = append(decoded_prefix1, decoded_prefix2...)
						//current_node.flag_value.encoded_prefix = append(current_node.flag_value.encoded_prefix, 16)
						//current_node.flag_value.value = node.flag_value.value
						combined_pref := append(decoded_prefix1, decoded_prefix2...)
						new_node := Node{2, [17]string{}, Flag_value{compact_encode(append(combined_pref, 16)), node.flag_value.value}} //convert to leaf node
						current_node = new_node
						//current_node.flag_value.encoded_prefix = compact_encode(append(compact_decode(node.flag_value.encoded_prefix),16))
					} else if compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 0 || compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 1 {
						fmt.Println("Entered two condition -- extension")
						decoded_prefix1 := compact_decode(current_node.flag_value.encoded_prefix)
						decoded_prefix2 := compact_decode(node.flag_value.encoded_prefix)
						current_node.flag_value.encoded_prefix = compact_encode(append(decoded_prefix1, decoded_prefix2...))
						//current_node.flag_value.encoded_prefix = compact_encode(compact_decode(node.flag_value.encoded_prefix))
					}
				}
			}
		}
	}
	if previous_node.hash_node() != current_node.hash_node() {
		fmt.Println("remove node :", previous_node)
		fmt.Println("hash of remove node :", previous_node.hash_node())
		if current_node.node_type == 0 {
			fmt.Println("ZERO NODE TYPE", " Node Hash: ", current_node.hash_node(), "current path: ", path)
			//remove zeroth node!
			//delete(mpt.db, current_node.hash_node())
		}
		delete(mpt.db, previous_node.hash_node())
		fmt.Println("add to mpt db, node value: ", current_node)
		fmt.Println("add to mpt db, node hash: ", current_node.hash_node())
		if current_node.node_type != 0 {
			mpt.db[current_node.hash_node()] = current_node
		}
	}
	return current_node
}

func compact_encode(hex_array []uint8) []uint8 {
	var term int
	if len(hex_array) == 0 { //HINT.TEST: Added to solve null path issue
		term = 0
	} else if hex_array[len(hex_array)-1] == 16 {
		term = 1
	} else {
		term = 0
	}
	if term == 1 {
		hex_array = hex_array[:len(hex_array)-1]
	}
	var oddlen int = len(hex_array) % 2
	var flags []uint8 = []uint8{uint8(2*term + oddlen)}
	if oddlen > 0 {
		hex_array = append(flags, hex_array...)
	} else {
		var zeroArr []uint8 = []uint8{0}
		flags = append(flags, zeroArr...)
		hex_array = append(flags, hex_array...)
	}
	var result []uint8
	fmt.Println(hex_array)
	for i := 0; i < len(hex_array); i += 2 {
		result = append(result, 16*hex_array[i]+hex_array[i+1])
	}
	fmt.Print(term, oddlen, flags)
	return result
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	result := compact_decode_wt_prefix(encoded_arr)
	//fmt.Println("result", result)
	//removed prefix from compact encode below
	firstNibble := result[0]
	if firstNibble == 0 || firstNibble == 2 {
		result = result[2:]
	} else {
		result = result[1:]
	}
	return result
}

func compact_decode_wt_prefix(encoded_arr []uint8) []uint8 {
	var result []uint8
	for i := 0; i < len(encoded_arr); i += 1 {
		result = append(result, encoded_arr[i]/16)
		result = append(result, encoded_arr[i]%16)
	}
	return result
}

//takes input of ascii and gives the prefix
func get_decode_prefix(ascii_arr []uint8) []uint8 {
	result := compact_decode_wt_prefix(ascii_arr)
	firstNibble := result[0]
	if firstNibble == 0 || firstNibble == 2 {
		//fmt.Println("Even first nibble")
		result = result[0:2]
	} else {
		//fmt.Println("Odd first nibble")
		result = result[0:1]
	}
	return result
}

func str_to_ascii(input string) []uint8 {
	if len(input) == 0 {
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
	return encoded_arr[0]/16 < 2
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
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
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
