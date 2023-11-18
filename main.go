package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Block struct {
	PrevHash  string
	Position  int
	Data      BookCheckout
	TimeStamp string
	Hash      string
}
type BookCheckout struct {
	BookId       string `json:"bookid"`
	User         string `json:"user"`
	CheckoutDate string `json:"checkoutdate"`
	Isgenesis    bool   `json:"isgenesis"`
}
type Book struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	PublishDate string `json:"publishdate"`
	ISBN        string `json:"isbn"`
}
type Blockchain struct {
	blocks []*Block
}

var BlockChain *Blockchain

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkout BookCheckout
	err := json.NewDecoder(r.Body).Decode(&checkout)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Change to BadRequest
		fmt.Println(err)
		w.Write([]byte("could not write block"))
		return
	}
	BlockChain.AddBlock(checkout)
}
func (bc *Blockchain) AddBlock(data BookCheckout) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	block := CreateBlock(prevBlock, data)

	if ValidBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	}
}
func CreateBlock(prevBlock *Block, checkout BookCheckout) *Block {
	block := &Block{}
	block.Position = prevBlock.Position + 1
	block.PrevHash = prevBlock.Hash
	block.TimeStamp = time.Now().String()
	block.GenerateHash()

	return block
}
func (b *Block) GenerateHash() {
	bytes, _ := json.Marshal(b.Data)
	data := string(rune(b.Position)) + b.TimeStamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}
func ValidBlock(block, perevBlock *Block) bool {
	if perevBlock.Hash != block.PrevHash {
		return false
	}
	if !block.ValidateHash(block.Hash) {
		return false
	}
	if perevBlock.Position+1 != block.Position {
		return false
	}
	return true
}
func (b *Block) ValidateHash(hash string) bool {
	if b.Hash != hash {
		return false
	}
	return true
}
func newBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write([]byte("could not create a new book"))
		return
	}
	h := md5.New()
	io.WriteString(h, book.ISBN+book.PublishDate)
	book.Id = fmt.Sprintf("%x", h.Sum(nil))
	resp, err := json.MarshalIndent(book, "", "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write([]byte("could not save book data"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
func NewBlockChain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}
func GenesisBlock() *Block {
	return CreateBlock(&Block{}, BookCheckout{Isgenesis: true})
}
func getBlockChain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	io.WriteString(w, string(jbytes))
}
func main() {
	BlockChain = NewBlockChain()
	r := mux.NewRouter()
	r.HandleFunc("/", getBlockChain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBook).Methods("POST")
	go func() {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev.hash:%x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", "")
			fmt.Printf("Data:%v\n", string(bytes))
			fmt.Printf("Hash%x\n", block.Hash)

		}
	}()
	fmt.Println("Listening New server")
	http.ListenAndServe(":3000", r)
}
