package lists

type ListElement struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// type ListData struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// 	Code string `json:"code"`
// 	Data *BinaryTree
// }

type List struct {
	ID                string             `json:"id"`
	Active            bool               `json:"active"`
	Name              string             `json:"name"`
	Code              string             `json:"code"`
	PaymentMediumCode *PaymentMediumType `json:"PaymentMediumType"`
	Version           float32            `json:"version"`
	OrganizationID    string             `json:"organizationId"`
	Metadata          *Metadata          `json:"metadata"`
	MediumIds         []string           `json:"mediumIds"`
	DataIds           *BinaryTree
	TimeUpload        int64
}

type PaymentMediumType struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IdRegex   string `json:"idRegex"`
	FlagRegex string `json:"flagRegex"`
	Active    bool   `json:"active"`
}

type Metadata struct {
	CreatedBy string `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	UpdatedBy string `json:"updatedBy"`
}

// {
//     "_id": "6b7c067b-8f58-45f1-b70c-a1cd402c26e5",
//     "active": true,
//     "name": "PRUEBA_NEBULAE",
//     "code": "LIST001",
//     "paymentMediumType": {
//         "id": "589c064b-7999-45aa-99e7-29a390f012f9",
//         "name": "T_M_P_NEBULA",
//         "idRegex": "[A-Z][0-9]",
//         "flagRegex": "g",
//         "active": true
//     },
//     "version": 0.1,
//     "organizationId": "604812bb-ed5a-4a46-b076-104dbf0dc982",
//     "metadata": {
//         "createdBy": "leonardo.gutierrez",
//         "createdAt": 1626376222010,
//         "updatedBy": "leonardo.gutierrez",
//         "updatedAt": 1626376222010
//     },
//     "mediumIds": [
//         "A77508451",
//         "A75084522",
//         "A75084523",
//         "A75084521"
//     ],
//     "id": "6b7c067b-8f58-45f1-b70c-a1cd402c26e5"
// }
