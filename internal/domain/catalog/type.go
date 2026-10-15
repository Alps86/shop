package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"time"
)

type AttributeCode = value.AttributeCode

type OptionCode = value.OptionCode

type OptionCodes []OptionCode

type AttributeType string

type AttributeLevel string

type LocalizedText map[value.Locale]string

type Attributes map[AttributeCode]AttributeValue

type AttributeDefinition struct {
	Code     AttributeCode
	Name     LocalizedText
	Type     AttributeType
	Level    AttributeLevel
	Unit     string
	Options  []AttributeOption
	Required bool
	Minimum  value.Optional[value.Decimal]
	Maximum  value.Optional[value.Decimal]
}

type AttributeOption struct {
	Code  OptionCode
	Label LocalizedText
}

type AttributeValue struct {
	Type        AttributeType
	Text        value.Optional[LocalizedText]
	Number      value.Optional[value.Decimal]
	Measurement value.Optional[Measurement]
	Options     []OptionCode
	Boolean     value.Optional[bool]
}

type Measurement struct {
	Value value.Decimal
	Unit  string
}

type Product struct {
	ID                    value.ProductID
	FamilyID              value.ProductFamilyID
	FamilyVersion         uint64
	Name                  LocalizedText
	Description           LocalizedText
	PrimaryCategoryID     value.CategoryID
	AdditionalCategoryIDs []value.CategoryID
	Attributes            Attributes
	Images                []Image
	Variants              []Variant
	Accessories           []AccessoryReference
	Brand                 string
	Manufacturer          EconomicOperator
	ResponsiblePerson     value.Optional[EconomicOperator]
	Compliance            ProductCompliance
	Revision              uint64
	Slug                  string
	Published             bool
}

type Variant struct {
	SKU                    value.SKU
	GTIN                   string
	ManufacturerPartNumber string
	Attributes             Attributes
	Images                 []Image
	Compliance             value.Optional[ProductCompliance]
	Manufacturer           value.Optional[EconomicOperator]
	ResponsiblePerson      value.Optional[EconomicOperator]
}

type Image struct {
	URL     string
	AltText LocalizedText
}

type AccessoryReference struct {
	ProductID      value.ProductID
	CompatibleSKUs []value.SKU
}

type ProductFamily struct {
	ID         value.ProductFamilyID
	Version    uint64
	Name       LocalizedText
	Attributes []AttributeDefinition
}

type Category struct {
	ID               value.CategoryID
	Version          uint64
	ParentID         value.Optional[value.CategoryID]
	Name             LocalizedText
	Description      LocalizedText
	FilterAttributes []AttributeCode
	TopAttributes    []AttributeCode
	Slug             string
}

type Catalog struct {
	Products   map[value.ProductID]Product
	Families   map[value.ProductFamilyID]ProductFamily
	Categories map[value.CategoryID]Category
}

type BusinessAddress struct {
	Street     string
	PostalCode string
	City       string
	Country    value.CountryCode
}

type EconomicOperator struct {
	Name              string
	TradeName         string
	PostalAddress     BusinessAddress
	ElectronicAddress string
	Phone             string
}

type ComplianceContext struct {
	Market             value.CountryCode
	Locale             value.Locale
	ValidFrom          value.Optional[time.Time]
	ValidUntil         value.Optional[time.Time]
	VerificationStatus string
}

type ProductCompliance struct {
	Context             ComplianceContext
	ApplicableRules     []RuleReference
	Safety              SafetyInformation
	ChemicalLabel       value.Optional[ChemicalLabel]
	EnergyLabel         value.Optional[EnergyLabel]
	NetContent          value.Optional[Measurement]
	UnitPriceBasis      value.Optional[Measurement]
	Documents           []ProductDocument
	Notices             []RequiredNotice
	RequiredFields      []string
	NonApplicableReason string
}

type RuleReference struct {
	Code       string
	Market     value.CountryCode
	ValidFrom  value.Optional[time.Time]
	ValidUntil value.Optional[time.Time]
}

type SafetyInformation struct {
	Warnings     []RequiredNotice
	Instructions []ProductDocument
}

type RequiredNotice struct {
	Code      string
	Rules     []RuleReference
	Text      LocalizedText
	Placement []DisplayContext
}

type DisplayContext string

type ChemicalLabel struct {
	ProductIdentifier       string
	Supplier                EconomicOperator
	NominalQuantity         value.Optional[Measurement]
	Pictograms              []HazardPictogram
	SignalWord              LocalizedText
	HazardStatements        []CodedStatement
	PrecautionaryStatements []CodedStatement
	SupplementalStatements  []CodedStatement
	UFI                     string
}

type HazardPictogram struct {
	Code     string
	ImageURL string
	AltText  LocalizedText
}

type CodedStatement struct {
	Code string
	Text LocalizedText
}

type EnergyLabel struct {
	SupplierModelIdentifier string
	EPRELRegistrationNumber string
	EPRELURL                string
	EfficiencyClass         string
	ScaleMin                string
	ScaleMax                string
	LabelDocuments          []ProductDocument
	InformationSheets       []ProductDocument
}

type ProductDocument struct {
	Kind     DocumentKind
	Locale   value.Locale
	SKU      value.SKU
	Model    string
	URL      string
	Version  string
	IssuedAt value.Optional[time.Time]
}

type DocumentKind string

type ResolvedCompliance struct {
	Manufacturer      EconomicOperator
	ResponsiblePerson value.Optional[EconomicOperator]
	Compliance        ProductCompliance
}
