package owl

// Ontology holds all information of one ontology
type Ontology struct {
	Prefixes            []Prefix              `xml:"Prefix"`
	Declarations        []Declaration         `xml:"Declaration"`
	SubClasses          []SubClassOf          `xml:"SubClassOf"`
	AnnotationAssertion []AnnotationAssertion `xml:"AnnotationAssertion"`
}

type Prefix struct {
	Name string `xml:"name,attr"`
	IRI  string `xml:"IRI,attr"`
}

type Declaration struct {
	Class           Class           `xml:"Class"`
	ObjectProperty  ObjectProperty  `xml:"ObjectProperty"`
	DataProperty    DataProperty    `xml:"DataProperty"`
	NamedIndividual NamedIndividual `xml:"NamedIndividual"`
}

type AnnotationAssertion struct {
	AnnotationProperty AnnotationProperty `xml:"AnnotationProperty"`
	IRI                string             `xml:"IRI"`
	AbbreviatedIRI     string             `xml:"AbbreviatedIRI"`
	Literal            string             `xml:"Literal"`
}

type Literal struct {
	Literal string `xml:"attr"`
}
type AnnotationProperty struct {
	AbbreviatedIRI string `xml:"abbreviatedIRI,attr"`
}

type Entity struct {
	IRI            string `xml:"IRI,attr"`
	AbbreviatedIRI string `xml:"abbreviatedIRI,attr"`
}

type Class struct {
	Entity
}

type ObjectProperty struct {
	Entity
}

type DataProperty struct {
	Entity
}

type NamedIndividual struct {
	Entity
}

type SubClassOf struct {
	Class                []Class                `xml:"Class"`
	ObjectSomeValuesFrom []ObjectSomeValuesFrom `xml:"ObjectSomeValuesFrom"`
	DataSomeValuesFrom   []DataSomeValuesFrom   `xml:"DataSomeValuesFrom"`
	ObjectHasValue       []ObjectHasValue       `xml:"ObjectHasValue"`
	DataHasValue         []DataHasValue         `xml:"DataHasValue"`
}

type ObjectSomeValuesFrom struct {
	ObjectProperty ObjectProperty `xml:"ObjectProperty"`
	Class          Class          `xml:"Class"`
}

type DataSomeValuesFrom struct {
	DataProperty DataProperty `xml:"DataProperty"`
	Datatype     Datatype     `xml:"Datatype"`
}

type DataHasValue struct {
	DataProperty DataProperty `xml:"DataProperty"`
	Literal      string       `xml:"Literal"`
}

type Datatype struct {
	AbbreviatedIRI string `xml:"abbreviatedIRI,attr"`
}

type ObjectHasValue struct {
	ObjectProperty  ObjectProperty  `xml:"ObjectProperty"`
	NamedIndividual NamedIndividual `xml:"NamedIndividual"`
}
