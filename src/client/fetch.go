package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"src/file"
	messagesStorage "src/messages/client_storage"
	proto3Storage "src/proto/client_storage"
	proto3 "src/proto/controller_client"
	"strings"
	"sync"
)

func (c *Client) FetchFile(res proto3.ResponseInterface) {
	c.logger.Info("Fetching file")
	fragments := res.(*proto3.FragLayoutResponse).FragmentLayout

	maxGoroutines := 10 // limit to 10 goroutines
	sem := make(chan struct{}, maxGoroutines)

	var wg sync.WaitGroup
	wg.Add(len(fragments))

	for _, frag := range fragments {
		go func(f proto3.FragmentInfo) {
			sem <- struct{}{} // acquire semaphore
			defer func() {
				<-sem // release semaphore
				wg.Done()
			}()
			c.FetchFragment(f)

		}(frag)
	}

	wg.Wait()
	c.logger.Info("All fragments fetched")
	c.logger.Info("Combining fragments")

	fileName := c.GetFileName(fragments[0].FragmentId)
	if fileName == "" {
		c.logger.Error("Error getting file name")
		return
	}

	//c.CombineFragments(file.DIR, "large-log.txt", len(fragments))
	err := c.CombineFragments(c.file.Dir(), fileName, len(fragments))
	if err != nil {
		c.logger.Error("Error combining fragments")
		return
	}

	c.logger.Info("File fetched and combined")

}

func (c *Client) GetFileName(fragID string) string {

	lastUnderscore := strings.LastIndex(fragID, "_")
	if lastUnderscore == -1 {
		// If there is no underscore, return the whole string
		return fragID
	}
	return fragID[:lastUnderscore]
}

func (c *Client) CombineFragments(dir string, file string, numFrags int) (err error) {
	// Create the output file
	outFile, err := os.Create(filepath.Join(dir, file))
	//outFile, err := os.Create(file)
	if err != nil {
		c.logger.Error("error creating output file")
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer func(outFile *os.File) {
		err = outFile.Close()
		if err != nil {
			c.logger.Error("error closing output file")
			return
		}
	}(outFile)

	// Read in the fragments and write them to the output file
	for i := 0; i < numFrags; i++ {
		// Construct the fragment file name

		fragFileName := fmt.Sprintf("%s_%d", file, i)
		c.logger.Sugar().Infof("Opening fragment file %s", fragFileName)

		// Open the fragment file
		fragFile, err := os.Open(filepath.Join(dir, fragFileName))
		if err != nil {
			err = fragFile.Close()
			if err != nil {
				c.logger.Error("error closing fragment file")
				return err
			}
			c.logger.Error("error opening fragment file")
			return fmt.Errorf("error opening fragment file %s: %v", fragFileName, err)
		}

		// Copy the fragment file to the output file
		c.logger.Sugar().Infof("Appending fragment %d to output file", i)
		_, err = io.Copy(outFile, fragFile)
		if err != nil {
			err := fragFile.Close()
			if err != nil {
				c.logger.Error("error closing fragment file")
				return err
			}
			return fmt.Errorf("error copying fragment %d to output file: %v", i, err)
		}
		err = fragFile.Close()
		if err != nil {
			c.logger.Error("error closing fragment file")
			return err
		}
	}

	return nil
}

func (c *Client) FetchFragment(frag proto3.FragmentInfo) {
	nodes := frag.StorageNodes

	for _, node := range nodes {
		err := c.FetchFromNode(frag, node)
		if err != nil {
			c.logger.Sugar().Error("There was an error fetching from node: ", node.NodeId)
			continue
		} else {
			break
		}
	}

}

func (c *Client) FetchFromNode(frag proto3.FragmentInfo, node proto3.StorageNodeInfo) (err error) {

	conn, err := net.Dial("tcp", node.Host+":"+node.Port)
	if err != nil {
		c.logger.Error("There was an error connecting to the host.")
		return
	}

	msgHandler := messagesStorage.NewMessageHandler(conn)
	proto := proto3Storage.NewProtoHandler(msgHandler, c.logger, c.file.Dir())

	fileHandler := file.FileHandler{}
	fileHandler.SetFileName(frag.FragmentId)
	fileHandler.SetDir(c.file.Dir())

	proto.SetFileHandler(&fileHandler)
	err = proto.HandleFileGetRequest(frag.FragmentId)
	if err != nil {
		c.logger.Error("There was an error handling the file get request.")
		return
	}

	c.handleStorageResponse(proto)
	return
}
