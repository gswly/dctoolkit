package main

import (
    dctk "github.com/gswly/dctoolkit"
)

// edit with the nick you want to download from and the file you want to download
const (
    PeerName = "othernick"
    FilePath = "/share/file.txt"
)

func main() {
    // automatically connect to hub. local ports must be opened and accessible (configure your router)
    client,err := dctk.NewClient(dctk.ClientConf{
        HubUrl: "nmdc://hubip:411",
        Nick: "mynick",
        TcpPort: 3005,
        UdpPort: 3005,
        TcpTlsPort: 3004,
    })
    if err != nil {
        panic(err)
    }

    // download peer file list
    client.OnPeerConnected = func(p *dctk.Peer) {
        if p.Nick == PeerName {
            client.DownloadFileList(p.Nick)
        }
    }

    filelistDownloaded := false
    client.OnDownloadSuccessful = func(d* dctk.Download) {
        // file list has been downloaded
        if filelistDownloaded == false {
            filelistDownloaded = true

            // parse file list
            fl,err := dctk.FileListParse(d.Content())
            if err != nil {
                panic(err)
            }

            // find file
            file,err := fl.GetFile(FilePath)
            if err != nil {
                panic(err)
            }

            // start downloading the file
            client.DownloadFLFile(d.Conf().Nick, file)

        // file has been downloaded
        } else {
            client.Terminate()
        }
    }

    client.Run()
}