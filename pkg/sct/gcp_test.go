package sct

type MyPayload struct {
	Foo int `json:"foo"`
}

//Removing this for now. We may want to set up testing credentials in future
/*

func TestGcpTokenCreate(t *testing.T) {
	c := context.Background()
	client, err := secretmanager.NewClient(c)
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()
	g := NewGcpTokenManager[MyPayload]("controlplane-test-268922", client, nil)
	defer g.Close()
	token, err := g.NewToken("sct.01234567890ABCDEF.01234567890124567890234567890123456789012345678901234567890", "dataServiceSecret")
	if err != nil {
		t.Error(err)
		panic(err)
	}
	token.Name = "some-name"

	//If the token already exists for some reason, this becomes an implicit test of the Delete function
	if e, err := g.Exists(c, token); err == nil && e {
		if err != nil {
			t.Error(err)
			panic(err)
		}
		if err = g.Delete(c, token); err != nil {
			t.Error(err)
			panic(err)
		}
	}

	err = g.Write(c, token)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = g.Delete(c, token)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func TestGcpTokenRead(t *testing.T) {
	c := context.Background()
	client, err := secretmanager.NewClient(c, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()
	g := NewGcpTokenManager[MyPayload]("controlplane-test-268922", client, nil)
	defer g.Close()

	//Ensure our token exists in the repository
	token, err := g.NewToken("sct.01234567890ABCDEF.01234567890124567890234567890123456789012345678901234567890", "dataServiceSecret")
	token.Content.Foo = 9001
	if err != nil {
		t.Error(err)
		return
	}
	token.Name = "some-name"

	err = g.Write(c, token)
	if err != nil {
		t.Error(err)
		return
	}

	//Read the token from the repository and compare it against the original
	readToken, err := g.NewToken("sct.01234567890ABCDEF.01234567890124567890234567890123456789012345678901234567890", "dataServiceSecret")
	err = g.Read(c, readToken)
	if err != nil {
		t.Error(err)
		return
	}

	if readToken.Content.Foo != token.Content.Foo {
		err = errors.New("repository did not preserve token content")
		t.Error(err)
		return
	}

	err = g.Delete(c, readToken)
	if err != nil {
		t.Error(err)
		return
	}
	return
}


*/
