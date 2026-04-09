package plc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

// ModbusClient wraps a Modbus TCP connection to a single PLC.
// Thread-safe: the underlying client is protected by a mutex.
type ModbusClient struct {
	client  *modbus.ModbusClient
	mu      sync.Mutex
	addr    string
	timeout time.Duration
}

// ModbusConfig holds connection parameters for a Modbus TCP client.
type ModbusConfig struct {
	Address string        // "host:port" (default port 502)
	Timeout time.Duration // per-operation timeout (default 10s)
}

// NewModbusClient creates a new Modbus TCP client for the given address.
// Does not connect immediately — call Connect() first.
func NewModbusClient(cfg ModbusConfig) (*ModbusClient, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     fmt.Sprintf("tcp://%s", cfg.Address),
		Timeout: cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create modbus client: %w", err)
	}

	return &ModbusClient{
		client:  client,
		addr:    cfg.Address,
		timeout: cfg.Timeout,
	}, nil
}

// Connect opens the TCP connection to the PLC.
func (c *ModbusClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.client.Open(); err != nil {
		return fmt.Errorf("modbus connect %s: %w", c.addr, err)
	}
	return nil
}

// Close closes the TCP connection.
func (c *ModbusClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.Close()
}

// Register represents a single Modbus register reading.
type Register struct {
	Address  uint16
	Value    uint16
	FloatVal float64
}

// ReadHoldingRegisters reads one or more holding registers starting at addr.
func (c *ModbusClient) ReadHoldingRegisters(_ context.Context, addr uint16, count uint16) ([]Register, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	regs, err := c.client.ReadRegisters(addr, count, modbus.HOLDING_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("read holding registers %d+%d: %w", addr, count, err)
	}

	result := make([]Register, len(regs))
	for i, v := range regs {
		result[i] = Register{
			Address:  addr + uint16(i),
			Value:    v,
			FloatVal: float64(v),
		}
	}
	return result, nil
}

// ReadInputRegisters reads one or more input registers starting at addr.
func (c *ModbusClient) ReadInputRegisters(_ context.Context, addr uint16, count uint16) ([]Register, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	regs, err := c.client.ReadRegisters(addr, count, modbus.INPUT_REGISTER)
	if err != nil {
		return nil, fmt.Errorf("read input registers %d+%d: %w", addr, count, err)
	}

	result := make([]Register, len(regs))
	for i, v := range regs {
		result[i] = Register{
			Address:  addr + uint16(i),
			Value:    v,
			FloatVal: float64(v),
		}
	}
	return result, nil
}

// WriteRegister writes a single value to a holding register.
func (c *ModbusClient) WriteRegister(_ context.Context, addr uint16, value uint16) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.client.WriteRegister(addr, value); err != nil {
		return fmt.Errorf("write register %d: %w", addr, err)
	}
	return nil
}

// ReadAndVerifyWrite writes a value, then reads it back to verify.
// Returns an error if the readback doesn't match.
func (c *ModbusClient) ReadAndVerifyWrite(ctx context.Context, addr uint16, value uint16) error {
	if err := c.WriteRegister(ctx, addr, value); err != nil {
		return err
	}

	regs, err := c.ReadHoldingRegisters(ctx, addr, 1)
	if err != nil {
		return fmt.Errorf("verify read after write: %w", err)
	}

	if len(regs) == 0 || regs[0].Value != value {
		actual := uint16(0)
		if len(regs) > 0 {
			actual = regs[0].Value
		}
		return fmt.Errorf("write verification failed: wrote %d, read back %d", value, actual)
	}

	return nil
}

// Addr returns the PLC address this client is connected to.
func (c *ModbusClient) Addr() string {
	return c.addr
}
