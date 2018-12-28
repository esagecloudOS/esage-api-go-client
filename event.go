package abiquo_api

type EventCollection struct {
	AbstractCollection
	Collection []Event
}

type Event struct {
	DTO
	ID                  int    `json:"id"`
	ActionPerformed     string `json:"actionPerformed"`
	Component           string `json:"component"`
	Datacenter          string `json:"datacenter"`
	Enterprise          string `json:"enterprise"`
	IDDatacenter        int    `json:"idDatacenter"`
	IDEnterprise        int    `json:"idEnterprise"`
	IDNetwork           int    `json:"idNetwork"`
	IDPhysicalMachine   int    `json:"idPhysicalMachine"`
	IDRack              int    `json:"idRack"`
	IDStoragePool       string `json:"idStoragePool"`
	IDStorageSystem     int    `json:"idStorageSystem"`
	IDSubnet            int    `json:"idSubnet"`
	IDUser              int    `json:"idUser"`
	IDVirtualApp        int    `json:"idVirtualApp"`
	IDVirtualDatacenter int    `json:"idVirtualDatacenter"`
	IDVirtualMachine    int    `json:"idVirtualMachine"`
	IDVolume            string `json:"idVolume"`
	Network             string `json:"network"`
	PerformedBy         string `json:"performedBy"`
	PhysicalMachine     string `json:"physicalMachine"`
	Rack                string `json:"rack"`
	Severity            string `json:"severity"`
	Stacktrace          string `json:"stacktrace"`
	StoragePool         string `json:"storagePool"`
	StorageSystem       string `json:"storageSystem"`
	Subnet              string `json:"subnet"`
	Timestamp           string `json:"timestamp"`
	User                string `json:"user"`
	VirtualApp          string `json:"virtualApp"`
	VirtualDatacenter   string `json:"virtualDatacenter"`
	VirtualMachine      string `json:"virtualMachine"`
	Volume              string `json:"volume"`
}
