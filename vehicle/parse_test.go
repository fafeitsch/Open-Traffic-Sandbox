package vehicle

import (
	"github.com/twpayne/go-polyline"
	"os"
	"testing"
)

func TestParseJson(t *testing.T) {
	jsonFile, err := os.Open("../samples/json/sample1.json")
	if err != nil {
		t.Errorf("Could not read the sample file: %v", err)
	}
	defer func() { _ = jsonFile.Close() }()
	input, err := parseJson(jsonFile)
	if err != nil {
		t.Errorf("Could not parse the sample file: %v", err)
	}
	vehicles := input.Vehicles
	expectedNumber := 2
	if len(vehicles) != expectedNumber {
		t.Errorf("%d vehicles should be extracted, but %d were.", expectedNumber, len(vehicles))
	}
	inifite := input.Settings.Infinite
	expectedInfinite := true
	if expectedInfinite != inifite {
		t.Errorf("Settings for inifinite sould be %t, but were %t.", expectedInfinite, inifite)
	}
}

const geometryVehicle2 = "qq}nH__}{@?c@s@{DKkAg@sEiA`Dc@fASCi@a@s@}@k@|Ae@`BIn@CV?VDz@LtAR|BD`@RbCPbBTpA\\|AF`@B`@?`@AXEVMb@ENWt@CLCVE?CBCBADAD?F@D@DBDB@D@BABCBE@ERFNLTZ|BvCZ^TH`Bl@Ix@YxBGbC?dA@zAZbJFpA^pKF\\JRpAFb@?tA?RArCQNCD?HAn@GnAMjBSJAD?HArASLCP?@vA?xBMnFKrGM~DEbAKvBAVI~AGtAn@Pr@VDFBRE~AGxAIlAmArVGhAAREn@?DQhDEx@GnACb@GfAIlAGnAGfAAZA`ABfA?t@E|@QtBEh@CNEh@CZCf@M|@Uv@QTMLQHU@UCOISOKQIUIWASCU?[DUFQLULIPALFh@`@j@r@h@f@TRRLj@Xd@Nb@LtBf@\\J\\JRFb@Nd@TnBjA\\T`@Xz@f@t@h@pD~DXZ|AzAnFhFbE`GT\\JNtDnHjArBb@t@^t@\\l@h@|@f@x@l@x@d@f@^\\f@b@j@d@f@\\d@Xd@X`@X^PdA\\pATpABlBCn@I^IbDq@tAYjIeBlDy@tCq@fA]tAm@lAk@LGPEVAR@TBRHPH`@TVNTJXF`@?^EfAUf@Ot@QXGn@Gj@Cx@?\\HVJTPVPXTdAhA|DnEBBr@t@^`@hBlBpDpDVTnDdDlAdAp@j@z@z@~ArA^~@H\\@R?VCXGTKNQPKJSAUESMKKOS[e@UWq@w@SWOSKY}@tAiB`D_CrD{AdCsCrEy@rAuBdDeGvNu@bBu@`AQNQNo@X\\v@JTHT`@fA`@jAPn@\\rANt@rAvFpArHJb@D`@PbALl@Rz@b@tAd@~@rBtC|AjDf@`B^zAH`@f@zBb@jCJjB?D?zEEhC?j@CpAEzBAN?x@CvBa@fE[pGKjGAv@@`G?jA?hAQdICxJBz@i@FKAKOIWK_BEWGAEF[bCGPGDGAEUO_D"

func routeService([]Coordinate) ([]Coordinate, float64, error) {
	bytes := []byte(geometryVehicle2)
	decoded, _, _ := polyline.DecodeCoords(bytes)
	coordinates := make([]Coordinate, 0, len(decoded))
	for _, c := range decoded {
		coordinates = append(coordinates, Coordinate{Lat: c[0], Lon: c[1]})
	}
	return coordinates, 10806, nil
}

func TestLoadRoutedVehicles(t *testing.T) {
	jsonFile, err := os.Open("../samples/json/sample1.json")
	if err != nil {
		t.Errorf("Could not read the sample file: %v", err)
	}
	defer func() { _ = jsonFile.Close() }()
	vehicles, err := LoadRoutedVehicles(jsonFile, routeService)
	if err != nil {
		t.Errorf("Could not create the RoutedVehicles: %v", err)
	}
	expectedLenght := 2
	if len(vehicles) != expectedLenght {
		t.Errorf("%d RoutedVehicles should have been created, but actually %d were extracted.", expectedLenght, len(vehicles))
	}
}
