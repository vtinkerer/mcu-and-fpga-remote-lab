package multiplexer

import "fmt"

type MultiplexerModule struct {
	mux1 *driverMultiplexer
	mux2 *driverMultiplexer
}

func NewMultiplexerModule(A0_1 int, A0_2 int, A1_1 int, A1_2 int) *MultiplexerModule {
	return &MultiplexerModule{
		mux1: newDriverMultiplexer(A0_1, A0_2),
		mux2: newDriverMultiplexer(A1_1, A1_2),
	}
}

func (m *MultiplexerModule) selectInputChannel(mux int, channel int) error {
	if mux == 1 {
		return m.mux1.selectInputChannel(channel)
	} else if mux == 2 {
		return m.mux2.selectInputChannel(channel)
	}
	return fmt.Errorf("invalid multiplexer: %d", mux)
}

func (m *MultiplexerModule) getInputChannel(mux int) (int, error) {
	if mux == 1 {
		return m.mux1.getInputChannel()
	} else if mux == 2 {
		return m.mux2.getInputChannel()
	}
	return 0, fmt.Errorf("invalid multiplexer: %d", mux)
}
