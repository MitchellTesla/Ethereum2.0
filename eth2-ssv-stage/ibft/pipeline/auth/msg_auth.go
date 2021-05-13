package auth

import (
	"errors"
	"github.com/bloxapp/ssv/ibft/pipeline"
	"github.com/bloxapp/ssv/ibft/proto"
)

// AuthorizeMsg is the pipeline to authorize message
func AuthorizeMsg(params *proto.InstanceParams) pipeline.Pipeline {
	return pipeline.WrapFunc("authorize", func(signedMessage *proto.SignedMessage) error {
		pks, err := params.PubKeysByID(signedMessage.SignerIds)
		if err != nil {
			return err
		}
		if len(pks) == 0 {
			return errors.New("could not find public key")
		}

		var foundVerified bool
		for _, pk := range pks {
			res, err := signedMessage.VerifySig(pk)
			if err != nil {
				return err
			}

			if res {
				foundVerified = true
				break
			}
		}
		if !foundVerified {
			return errors.New("could not verify message signature")
		}

		return nil
	})
}
