package updates

import (
	"golang.org/x/xerrors"
)

func validatePts(pts, ptsCount int) error {
	if pts < 0 {
		return xerrors.Errorf("invalid pts value: %d", pts)
	}

	if ptsCount < 0 {
		return xerrors.Errorf("invalid ptsCount value: %d", ptsCount)
	}

	return nil
}

func validateQts(qts int) error {
	if qts < 0 {
		return xerrors.Errorf("invalid qts value: %d", qts)
	}

	return nil
}

func validateSeq(seq, seqStart int) error {
	if seq < 0 {
		return xerrors.Errorf("invalid seq value: %d", seq)
	}

	if seqStart < 0 {
		return xerrors.Errorf("invalid seqStart value: %d", seq)
	}

	return nil
}
