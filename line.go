package xinvoice

import (
	"strconv"

	"github.com/invopop/gobl/bill"
)

// Line defines the structure of the IncludedSupplyChainTradeLineItem in the CII standard
type Line struct {
	ID              string           `xml:"ram:AssociatedDocumentLineDocument>ram:LineID"`
	Name            string           `xml:"ram:SpecifiedTradeProduct>ram:Name"`
	NetPrice        string           `xml:"ram:SpecifiedLineTradeAgreement>ram:NetPriceProductTradePrice>ram:ChargeAmount"`
	TradeDelivery   *Quantity        `xml:"ram:SpecifiedLineTradeDelivery>ram:BilledQuantity"`
	TradeSettlement *TradeSettlement `xml:"ram:SpecifiedLineTradeSettlement"`
}

// Quantity defines the structure of the quantity with its attributes for the CII standard
type Quantity struct {
	Amount   string `xml:",chardata"`
	UnitCode string `xml:"unitCode,attr"`
}

// TradeSettlement defines the structure of the SpecifiedLineTradeSettlement of the CII standard
type TradeSettlement struct {
	ApplicableTradeTax            []*ApplicableTradeTax            `xml:"ram:ApplicableTradeTax"`
	SpecifiedTradeAllowanceCharge []*SpecifiedTradeAllowanceCharge `xml:"ram:SpecifiedTradeAllowanceCharge"`
	Sum                           string                           `xml:"ram:SpecifiedTradeSettlementLineMonetarySummation>ram:LineTotalAmount"`
}

// ApplicableTradeTax defines the structure of ApplicableTradeTax of the CII standard
type ApplicableTradeTax struct {
	TaxType        string `xml:"ram:TypeCode"`
	TaxCode        string `xml:"ram:CategoryCode"`
	TaxRatePercent string `xml:"ram:RateApplicablePercent"`
}

// SpecifiedTradeAllowanceCharge defines the structure of SpecifiedTradeAllowanceCharge of the CII standard
type SpecifiedTradeAllowanceCharge struct {
	Indicator bool   `xml:"ram:ChargeIndicator>udt:Indicator"`
	Amount    string `xml:"ram:ActualAmount"`
	Code      string `xml:"ram:ReasonCode,omitempty"`
	Reason    string `xml:"ram:Reason,omitempty"`
}

func newLine(line *bill.Line) *Line {
	if line.Item == nil {
		return nil
	}
	item := line.Item

	lineItem := &Line{
		ID:       strconv.Itoa(line.Index),
		Name:     item.Name,
		NetPrice: item.Price.String(),
		TradeDelivery: &Quantity{
			Amount:   line.Quantity.String(),
			UnitCode: string(item.Unit.UNECE()),
		},
		TradeSettlement: newTradeSettlement(line),
	}

	return lineItem
}

func newTradeSettlement(line *bill.Line) *TradeSettlement {
	var applicableTradeTax []*ApplicableTradeTax
	var specifiedTradeAllowanceCharge []*SpecifiedTradeAllowanceCharge
	for _, tax := range line.Taxes {
		tradeTax := &ApplicableTradeTax{
			TaxType: tax.Category.String(),
			TaxCode: FindTaxCode(tax.Rate),
		}

		if tax.Percent != nil {
			tradeTax.TaxRatePercent = tax.Percent.StringWithoutSymbol()
		}

		applicableTradeTax = append(applicableTradeTax, tradeTax)
	}
	for _, discount := range line.Discounts {
		tradeAllowanceChange := &SpecifiedTradeAllowanceCharge{
			Indicator: false,
			Amount:    discount.Amount.String(),
			Reason:    discount.Reason,
			Code:      string(discount.Code),
		}

		specifiedTradeAllowanceCharge = append(specifiedTradeAllowanceCharge, tradeAllowanceChange)
	}
	for _, charge := range line.Charges {
		tradeAllowanceChange := &SpecifiedTradeAllowanceCharge{
			Indicator: true,
			Amount:    charge.Amount.String(),
			Reason:    charge.Reason,
			Code:      string(charge.Code),
		}

		specifiedTradeAllowanceCharge = append(specifiedTradeAllowanceCharge, tradeAllowanceChange)
	}

	settlement := &TradeSettlement{
		ApplicableTradeTax:            applicableTradeTax,
		SpecifiedTradeAllowanceCharge: specifiedTradeAllowanceCharge,
		Sum:                           line.Total.String(),
	}

	return settlement
}

// NewLines generates lines for XInvoice
func NewLines(lines []*bill.Line) []*Line {
	var Lines []*Line

	for _, line := range lines {
		Lines = append(Lines, newLine(line))
	}

	return Lines
}
