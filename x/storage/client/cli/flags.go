package cli

const (
	FlagPublic         = "public"
	FlagPaymentAccount = "payment-account"
	// TODO: Use a primary-account instead, which can load account from keyring and sign automatically.
	FlagPrimarySPApproval = "primary-sp-signature"
)
