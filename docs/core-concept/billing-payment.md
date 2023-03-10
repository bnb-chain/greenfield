# Billing and Payment

Greenfield will charge the users in two parts. Firstly, every
transaction will require gas fees to pay the Greenfield validator to
write the metadata on-chain. Secondly, the SPs charge the users for
their storage service. Such payment also happens on the Greenfield. This
document is about the latter: how such off-chain service fees are billed
and charged.

There are two kinds of fees for the off-chain service: object storage
fee and read fee:

1. Every object stored on Greenfield is charged a fee based on its
   size. The storage price is determined by the service providers.

2. There is a free quota for users' objects to be read based on their
   size, content types, and more. If exceeded, i.e. the object data
   has been downloaded too many times, SP will limit the bandwidth
   for more downloads. Users can raise their read quota to
   get more download quota. The read quota price is determined by the
   Primary Storage Provider users selected.

The fees are paid on Greenfield in the style of
"Stream" from users to the SPs at a constant rate. The fees are charged
every second as they are used.

For more tech details, please refer to the [stream payment module design](../modules/billing_and_payment.md).