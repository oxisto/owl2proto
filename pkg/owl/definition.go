package owl

// TODOs
// Add annotationAssertion, e.g. label, comment,...

// Ontology holds all information of one ontology
type Ontology struct {
	// Declarations []Declaration `xml:"Declaration"`
	SubClasses []SubClassOf `xml:"SubClassOf"`
}

type Declaration struct {
	Class           Class           `xml:"Class"`
	ObjectProperty  ObjectProperty  `xml:"ObjectProperty"`
	DataProperty    DataProperty    `xml:"DataProperty"`
	NamedIndividual NamedIndividual `xml:"NamedIndividual"`
}

type Class struct {
	IRI string `xml:"IRI,attr"`
}

type ObjectProperty struct {
	AbbreviatedIRI string `xml:"abbreviatedIRI,attr"`
	IRI            string `xml:"IRI,attr"`
}

type DataProperty struct {
	AbbreviatedIRI string `xml:"abbreviatedIRI,attr"`
	IRI            string `xml:"IRI,attr"`
}

type DataType struct {
	AbbreviatedIRI string `xml:"abbreviatedIRI,attr"`
	IRI            string `xml:"IRI,attr"`
}

type NamedIndividual struct {
	AbbreviatedIRI string `xml:"abbreviatedIRI,attr"`
	IRI            string `xml:"IRI,attr"`
}

type SubClassOf struct {
	Class                []Class                `xml:"Class"`
	ObjectSomeValuesFrom []ObjectSomeValuesFrom `xml:"ObjectSomeValuesFrom"`
	DataSomeValuesFrom   []DataSomeValuesFrom   `xml:"DataSomeValuesFrom"`
	ObjectHasValue       []ObjectHasValue       `xml:"ObjectHasValue"`
}

type ObjectSomeValuesFrom struct {
	ObjectProperty string  `xml:"ObjectProperty"`
	Class          []Class `xml:"Class"`
}

type DataSomeValuesFrom struct {
	DataProperty string     `xml:"DataProperty"`
	DataType     []DataType `xml:"DataType"`
}

type ObjectHasValue struct {
	ObjectProperty string  `xml:"ObjectProperty"`
	Class          []Class `xml:"Class"`
}
