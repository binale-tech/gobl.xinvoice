package xinvoice

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/tax"
)

// Settlement defines the structure of ApplicableHeaderTradeSettlement of the CII standard
type Settlement struct {
	Currency           string              `xml:"ram:InvoiceCurrencyCode"`
	TypeCode           string              `xml:"ram:SpecifiedTradeSettlementPaymentMeans>ram:TypeCode"`
	Tax                []*Tax              `xml:"ram:ApplicableTradeTax"`
	PaymentTerms       string              `xml:"ram:SpecifiedTradePaymentTerms>ram:Description,omitempty"`
	Summary            *Summary            `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
	ReferencedDocument *ReferencedDocument `xml:"ram:InvoiceReferencedDocument,omitempty"`
}

// Tax defines the structure of ApplicableTradeTax of the CII standard
type Tax struct {
	CalculatedAmount      string `xml:"ram:CalculatedAmount"`
	TypeCode              string `xml:"ram:TypeCode"`
	BasisAmount           string `xml:"ram:BasisAmount"`
	CategoryCode          string `xml:"ram:CategoryCode"`
	RateApplicablePercent string `xml:"ram:RateApplicablePercent"`
}

// Summary defines the structure of SpecifiedTradeSettlementHeaderMonetarySummation of the CII standard
type Summary struct {
	TotalAmount          string          `xml:"ram:LineTotalAmount"`
	ChargeTotalAmount    string          `xml:"ram:ChargeTotalAmount,omitempty"`
	AllowanceTotalAmount string          `xml:"ram:AllowanceTotalAmount,omitempty"`
	TaxBasisTotalAmount  string          `xml:"ram:TaxBasisTotalAmount"`
	TaxTotalAmount       *TaxTotalAmount `xml:"ram:TaxTotalAmount"`
	GrandTotalAmount     string          `xml:"ram:GrandTotalAmount"`
	DuePayableAmount     string          `xml:"ram:DuePayableAmount"`
}

// ReferencedDocument defines the structure of InvoiceReferencedDocument of the CII standard
type ReferencedDocument struct {
	IssuerAssignedID string `xml:"ram:IssuerAssignedID,omitempty"`
	IssueDate        *Date  `xml:"ram:FormattedIssueDateTime>qdt:DateTimeString,omitempty"`
}

// TaxTotalAmount defines the structure of the TaxTotalAmount of the CII standard
type TaxTotalAmount struct {
	Amount   string `xml:",chardata"`
	Currency string `xml:"currencyID,attr"`
}

// NewSettlement creates the ApplicableHeaderTradeSettlement part of a EN 16931 compliant invoice
func NewSettlement(inv *bill.Invoice) *Settlement {
	settlement := &Settlement{
		Currency: string(inv.Currency),
		TypeCode: FindTypeCode(inv),
	}
	if inv.Payment != nil && inv.Payment.Terms != nil {
		settlement.PaymentTerms = inv.Payment.Terms.Detail
	}

	if inv.Totals != nil {
		settlement.Tax = newTaxes(inv.Totals.Taxes)
		settlement.Summary = newSummary(inv.Totals, string(inv.Currency))
	}

	if inv.Preceding != nil && len(inv.Preceding) > 0 {
		cor := inv.Preceding[0]
		settlement.ReferencedDocument = &ReferencedDocument{
			IssuerAssignedID: invoiceNumber(cor.Series, cor.Code),
			IssueDate: &Date{
				Date:   formatIssueDate(*cor.IssueDate),
				Format: "102",
			},
		}
	}

	return settlement
}

func newSummary(totals *bill.Totals, currency string) *Summary {
	summary := &Summary{
		TotalAmount:         totals.Total.String(),
		TaxBasisTotalAmount: totals.Total.String(),
		GrandTotalAmount:    totals.TotalWithTax.String(),
		DuePayableAmount:    totals.Payable.String(),
		TaxTotalAmount: &TaxTotalAmount{
			Amount:   totals.Tax.String(),
			Currency: currency,
		},
	}
	if totals.Discount != nil {
		summary.AllowanceTotalAmount = totals.Discount.String()
	}
	if totals.Charge != nil {
		summary.ChargeTotalAmount = totals.Charge.String()
	}
	return summary
}

func newTaxes(total *tax.Total) []*Tax {
	var Taxes []*Tax

	if total == nil {
		return nil
	}

	for _, category := range total.Categories {
		for _, rate := range category.Rates {
			tax := newTax(rate, category)

			Taxes = append(Taxes, tax)
		}
	}

	return Taxes
}

func newTax(rate *tax.RateTotal, category *tax.CategoryTotal) *Tax {
	if rate.Percent == nil {
		return nil
	}

	tax := &Tax{
		CalculatedAmount:      rate.Amount.String(),
		TypeCode:              category.Code.String(),
		BasisAmount:           rate.Base.String(),
		CategoryCode:          FindTaxCode(rate.Key),
		RateApplicablePercent: rate.Percent.StringWithoutSymbol(),
	}

	return tax
}
