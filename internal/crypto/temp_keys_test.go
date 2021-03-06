package crypto

import (
	"bytes"
	"crypto/aes"
	"math/big"
	"testing"

	"github.com/gotd/ige"
)

func hexInt(hexValue string) *big.Int {
	n, ok := big.NewInt(0).SetString(hexValue, 16)
	if !ok {
		panic(ok)
	}
	return n
}

func TestTempAESKeys(t *testing.T) {
	// https://core.telegram.org/mtproto/samples-auth_key#conversion-of-encrypted-answer-into-answer
	var (
		newNonce    = hexInt("311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D")
		serverNonce = hexInt("A5CF4D33F4A11EA877BA4AA573907330")
		keyExpected = hexInt("F011280887C7BB01DF0FC4E17830E0B91FBB8BE4B2267CB985AE25F33B527253")
		ivExpected  = hexInt("3212D579EE35452ED23E0D0C92841AA7D31B2E9BDEF2151E80D15860311C85DB")
	)
	key, iv := TempAESKeys(newNonce, serverNonce)
	if !bytes.Equal(key, keyExpected.Bytes()) {
		t.Error("invalid key")
	}
	if !bytes.Equal(iv, ivExpected.Bytes()) {
		t.Error("invalid iv")
	}

	encryptedAnswer := hexInt("28A92FE20173B347A8BB324B5FAB2667C9A8BBCE6468D5B509A4CB" +
		"DDC186240AC912CF7006AF8926DE606A2E74C0493CAA57741E6C82451F54D3E068F5CCC49B4444124B966" +
		"6FFB405AAB564A3D01E67F6E912867C8D20D9882707DC330B17B4E0DD57CB53BFAAFA9EF5BE76AE6C1B9B6" +
		"C51E2D6502A47C883095C46C81E3BE25F62427B585488BB3BF239213BF48EB8FE34C9A026CC8413934043" +
		"974DB03556633038392CECB51F94824E140B98637730A4BE79A8F9DAFA39BAE81E1095849EA4C83467C9" +
		"2A3A17D997817C8A7AC61C3FF414DA37B7D66E949C0AEC858F048224210FCC61F11C3A910B431CCBD104" +
		"CCCC8DC6D29D4A5D133BE639A4C32BBFF153E63ACA3AC52F2E4709B8AE01844B142C1EE89D075D64F69A" +
		"399FEB04E656FE3675A6F8F412078F3D0B58DA15311C1A9F8E53B3CD6BB5572C294904B726D0BE337E2E2" +
		"1977DA26DD6E33270251C2CA29DFCC70227F0755F84CFDA9AC4B8DD5F84F1D1EB36BA45CDDC70444D8C21" +
		"3E4BD8F63B8AB95A2D0B4180DC91283DC063ACFB92D6A4E407CDE7C8C69689F77A007441D4A6A8384B666" +
		"502D9B77FC68B5B43CC607E60A146223E110FCB43BC3C942EF981930CDC4A1D310C0B64D5E55D308D86325" +
		"1AB90502C3E46CC599E886A927CDA963B9EB16CE62603B68529EE98F9F5206419E03FB458EC4BD9454AA8F6" +
		"BA777573CC54B328895B1DF25EAD9FB4CD5198EE022B2B81F388D281D5E5BC580107CA01A50665C32B55271" +
		"5F335FD76264FAD00DDD5AE45B94832AC79CE7C511D194BC42B70EFA850BB15C2012C5215CABFE97CE66B8D8" +
		"734D0EE759A638AF013").Bytes()
	cipher, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	d := ige.NewIGEDecrypter(cipher, iv)
	decrypted := make([]byte, len(encryptedAnswer))
	d.CryptBlocks(decrypted, encryptedAnswer)

	expectedAnswer := hexInt("BA0D89B53E0549828CCA27E966B301A48FECE2FCA5CF4D33F4A1" +
		"1EA877BA4AA57390733002000000FE000100C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E2" +
		"1C14934D037563D930F48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C3" +
		"720FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F642477FE96BB2A941D5" +
		"BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4A4A695811051907E162753B56B0F6B410DBA" +
		"74D8A84B2A14B3144E0EF1284754FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA392" +
		"8FEF5B9AE4E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F0D8115F63" +
		"5B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5BFE000100262AABA621CC4DF587D" +
		"C94CF8252258C0B9337DFB47545A49CDD5C9B8EAE7236C6CADC40B24E88590F1CC2CC762EBF1CF11DC" +
		"C0B393CAAD6CEE4EE5848001C73ACBB1D127E4CB93072AA3D1C8151B6FB6AA6124B7CD782EAF981BDCF" +
		"CE9D7A00E423BD9D194E8AF78EF6501F415522E44522281C79D906DDB79C72E9C63D83FB2A940FF77" +
		"9DFB5F2FD786FB4AD71C9F08CF48758E534E9815F634F1E3A80A5E1C2AF210C5AB762755AD4B2126DF" +
		"A61A77FA9DA967D65DFD0AFB5CDF26C4D4E1A88B180F4E0D0B45BA1484F95CB2712B50BF3F5968D9D5" +
		"5C99C0FB9FB67BFF56D7D4481B634514FBA3488C4CDA2FC0659990E8E868B28632875A9AA703BCDCE8FCB7AE551").Bytes()

	answer := GuessDataWithHash(decrypted)
	if !bytes.Equal(answer, expectedAnswer) {
		t.Fatal("mismatch")
	}
}
