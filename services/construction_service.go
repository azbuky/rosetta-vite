package services

import (
	"context"
	"fmt"

	"github.com/azbuky/rosetta-vite/configuration"
	"github.com/azbuky/rosetta-vite/vite"
	"github.com/coinbase/rosetta-sdk-go/types"

	viteTypes "github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"github.com/vitelabs/go-vite/ledger"
)

// ConstructionAPIService implements the server.ConstructionAPIServicer interface.
type ConstructionAPIService struct {
	config *configuration.Configuration
	client Client
}

// NewConstructionAPIService creates a new instance of a ConstructionAPIService.
func NewConstructionAPIService(
	cfg *configuration.Configuration,
	client Client,
) *ConstructionAPIService {
	return &ConstructionAPIService{
		config: cfg,
		client: client,
	}
}

// ConstructionDerive implements the /construction/derive endpoint.
func (s *ConstructionAPIService) ConstructionDerive(
	ctx context.Context,
	request *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {

	addr := viteTypes.PubkeyToAddress(request.PublicKey.Bytes)
	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: addr.Hex(),
		},
	}, nil
}

// ConstructionPreprocess implements the /construction/preprocess
// endpoint.
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	options, requiredPublicKeys, err := vite.ConstructionPreprocess(request.Operations, request.Metadata)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	marshaledOptions, err := marshalJSONMap(options)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	return &types.ConstructionPreprocessResponse{
		Options:            marshaledOptions,
		RequiredPublicKeys: requiredPublicKeys,
	}, nil
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	if s.config.Mode != configuration.Online {
		return nil, ErrUnavailableOffline
	}

	var options vite.ConstructionOptions
	if err := unmarshalJSONMap(request.Options, &options); err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	metadata, err := s.client.ConstructionMetadata(ctx, &options)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	metadataMap, err := marshalJSONMap(metadata)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	return &types.ConstructionMetadataResponse{
		Metadata: metadataMap,
	}, nil
}

// ConstructionPayloads implements the /construction/payloads endpoint.
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {

	// Convert map to Metadata struct
	metadata := &vite.ConstructionMetadata{}
	if err := unmarshalJSONMap(request.Metadata, metadata); err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	description, err := vite.MatchTransaction(request.Operations)
	if err != nil {
		return nil, wrapErr(ErrUnclearIntent, err)
	}
	if len(request.PublicKeys) < 1 {
		return nil, wrapErr(ErrUnclearIntent, fmt.Errorf("missing public key"))
	}
	publicKey := request.PublicKeys[0]

	accountBlock, err := vite.CreateAccountBlock(description, metadata, publicKey)
	if err != nil {
		return nil, wrapErr(ErrUnclearIntent, err)
	}

	hash, err := accountBlock.ComputeHash()
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	// Construct SigningPayload
	payload := &types.SigningPayload{
		AccountIdentifier: &description.Account,
		Bytes:             hash.Bytes(),
		SignatureType:     types.Ed25519,
	}

	unsignedTransaction, err := encodeAccountBlockToBase64(*accountBlock)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: unsignedTransaction,
		Payloads:            []*types.SigningPayload{payload},
	}, nil
}

// ConstructionCombine implements the /construction/combine
// endpoint.
func (s *ConstructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	accountBlock, err := decodeAccountBlockFromBase64(request.UnsignedTransaction)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	// Compute hash for block
	hash, err := accountBlock.ComputeHash()
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	signature := request.Signatures[0]
	publicKey := ed25519.PublicKey(signature.PublicKey.Bytes)

	// Check signature
	ok := ed25519.Verify(publicKey, hash.Bytes(), signature.Bytes)
	if !ok {
		return nil, wrapErr(ErrSignatureInvalid, fmt.Errorf("signature is invalid"))
	}

	accountBlock.Signature = signature.Bytes
	accountBlock.PublicKey = signature.PublicKey.Bytes
	accountBlock.Hash = *hash

	signedTransaction, err := encodeAccountBlockToBase64(*accountBlock)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: signedTransaction,
	}, nil
}

// ConstructionHash implements the /construction/hash endpoint.
func (s *ConstructionAPIService) ConstructionHash(
	ctx context.Context,
	request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	accountBlock, err := decodeAccountBlockFromBase64(request.SignedTransaction)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	hash := accountBlock.Hash.Hex()

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: hash,
		},
	}, nil
}

// ConstructionParse implements the /construction/parse endpoint.
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {

	accountBlock, err := decodeAccountBlockFromBase64(request.Transaction)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	ops, err := vite.OperationsForAccountBlock(accountBlock, false)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	metadata := map[string]interface{}{
		"height":       accountBlock.Height,
		"previousHash": accountBlock.PrevHash,
		"difficulty":   accountBlock.Difficulty,
		"nonce":        accountBlock.Nonce,
		"blockType":    accountBlock.BlockType,
		"fee":          accountBlock.Fee,
		"data":         accountBlock.Data,
	}
	if ledger.IsReceiveBlock(accountBlock.BlockType) {
		metadata["sendBlockHash"] = accountBlock.SendBlockHash
	}

	resp := &types.ConstructionParseResponse{
		Operations: ops,
		Metadata:   metadata,
	}
	if request.Signed {
		resp.AccountIdentifierSigners = []*types.AccountIdentifier{
			{
				Address: accountBlock.Address.Hex(),
			},
		}
	}
	return resp, nil
}

// ConstructionSubmit implements the /construction/submit endpoint.
func (s *ConstructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	if s.config.Mode != configuration.Online {
		return nil, ErrUnavailableOffline
	}

	accountBlock, err := decodeAccountBlockFromBase64(request.SignedTransaction)
	if err != nil {
		return nil, wrapErr(ErrUnableToParseIntermediateResult, err)
	}

	if ledger.IsReceiveBlock(accountBlock.BlockType) {
		accountBlock.TokenId = viteTypes.ZERO_TOKENID
		accountBlock.Amount = nil
		accountBlock.Fee = nil
		accountBlock.ToAddress = viteTypes.ZERO_ADDRESS
	}

	if err := s.client.SendTransaction(ctx, accountBlock); err != nil {
		return nil, wrapErr(ErrBroadcastFailed, err)
	}

	txIdentifier := &types.TransactionIdentifier{
		Hash: accountBlock.Hash.Hex(),
	}
	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: txIdentifier,
	}, nil
}
