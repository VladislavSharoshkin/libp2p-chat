package internal

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"libp2p-chat/gen/model"
	. "libp2p-chat/gen/table"
	"libp2p-chat/misk"
)

func NewPeer(ID string, AddrInfo string) model.Peers {
	return model.Peers{ID: ID, AddrInfo: AddrInfo}
}

func SavePeer(pi peer.AddrInfo) (model.Peers, error) {

	//pk, err := pi.ID.ExtractPublicKey()
	//if err != nil {
	//	return piS, err
	//}
	//pkByte, err := crypto.MarshalPublicKey(pk)
	//if err != nil {
	//	return piS, err
	//}
	json, err := pi.MarshalJSON()
	if err != nil {
		return model.Peers{}, err
	}

	piS := NewPeer(pi.ID.String(), misk.ToBase64(json))
	insertStmt := Peers.
		INSERT(Peers.AllColumns).
		MODEL(piS).
		RETURNING(Peers.AllColumns)

	err = insertStmt.Query(DB, &piS)

	return piS, err
}
