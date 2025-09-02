package upload

import "os"


func Upload(filepath string, content []byte) (stdout []byte, stderr []byte, err error) {

    f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if _, err := f.Write(content); err != nil {
        return []byte{}, []byte{}, err
    }
    return []byte("File uploaded to " + filepath), nil, nil

}
