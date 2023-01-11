package keeper

//func TestGetBNBPrice(t *testing.T) {
//	type args struct {
//		priceTime int64
//	}
//	tests := []struct {
//		name          string
//		args          args
//		wantNum       math.Int
//		wantPrecision math.Int
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotNum, gotPrecision := GetBNBPrice(tt.args.priceTime)
//			if !reflect.DeepEqual(gotNum, tt.wantNum) {
//				t.Errorf("GetBNBPrice() gotNum = %v, want %v", gotNum, tt.wantNum)
//			}
//			if !reflect.DeepEqual(gotPrecision, tt.wantPrecision) {
//				t.Errorf("GetBNBPrice() gotPrecision = %v, want %v", gotPrecision, tt.wantPrecision)
//			}
//		})
//	}
//}
//
//func TestGetReadPrice(t *testing.T) {
//	type args struct {
//		readPacket types.ReadPacket
//		priceTime  int64
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    math.Int
//		wantErr bool
//	}{
//		{"zero", args{types.ReadPacketLevelFree, 0}, math.ZeroInt(), false},
//		{"0.1 USD", args{types.ReadPacketLevel1GB, 0}, math.NewInt(360490266762797), false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := GetReadPrice(tt.args.readPacket, tt.args.priceTime)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetReadPrice() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetReadPrice() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestGetReadPriceV0(t *testing.T) {
//	type args struct {
//		readPacket types.ReadPacket
//	}
//	tests := []struct {
//		name      string
//		args      args
//		wantPrice math.Int
//		wantErr   bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotPrice, err := GetReadPriceV0(tt.args.readPacket)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetReadPriceV0() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(gotPrice, tt.wantPrice) {
//				t.Errorf("GetReadPriceV0() gotPrice = %v, want %v", gotPrice, tt.wantPrice)
//			}
//		})
//	}
//}
//
//func TestSubmitBNBPrice(t *testing.T) {
//	type args struct {
//		priceTime int64
//		price     math.Int
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			SubmitBNBPrice(tt.args.priceTime, tt.args.price)
//		})
//	}
//}
