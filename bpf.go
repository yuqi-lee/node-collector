package main

type Key struct {
	Src      int32
	Dst      int32
	Protocol int32
}

type Value struct {
	Packets int64
	Bytes   int64
}

/*
func bpfMapLoad() {
	m, err := ebpf.LoadPinnedMap(path, nil)
	if err != nil {
		panic(err)
	} else {
		log.Printf("bpf map is loaded successfully with key size %d bytes.", m.KeySize())
	}
}
*/
