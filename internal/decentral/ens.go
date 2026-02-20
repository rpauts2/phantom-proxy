package decentral

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// ENSClient клиент для ENS
type ENSClient struct {
	mu            sync.RWMutex
	logger        *zap.Logger
	config        *ENSConfig
	client        *ethclient.Client
	privateKey    *ecdsa.PrivateKey
	publicAddress common.Address
	ensNames      map[string]string // name -> content hash
}

// ENSConfig конфигурация ENS
type ENSConfig struct {
	// Ethereum RPC
	RPCURL string
	
	// Приватный ключ для транзакций
	PrivateKey string
	
	// Контракты
	ENSRegistryAddress string
	ResolverAddress    string
	
	// Gas настройки
	GasPrice  *big.Int
	GasLimit  uint64
}

// NewENSClient создаёт новый ENS клиент
func NewENSClient(config *ENSConfig, logger *zap.Logger) (*ENSClient, error) {
	if config.RPCURL == "" {
		config.RPCURL = "https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
	}
	
	// Подключение к Ethereum
	client, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum: %w", err)
	}
	
	// Парсинг приватного ключа
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(config.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	
	publicAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	
	c := &ENSClient{
		logger:        logger,
		config:        config,
		client:        client,
		privateKey:    privateKey,
		publicAddress: publicAddress,
		ensNames:      make(map[string]string),
	}
	
	c.logger.Info("ENS client initialized",
		zap.String("address", publicAddress.Hex()))
	
	return c, nil
}

// RegisterENS регистрирует ENS имя
func (c *ENSClient) RegisterENS(ctx context.Context, name string, contentHash string) error {
	c.logger.Info("Registering ENS name",
		zap.String("name", name),
		zap.String("content_hash", contentHash))
	
	// TODO: Интеграция с ENS Registry контрактом
	// Для простоты симулируем
	
	c.mu.Lock()
	c.ensNames[name] = contentHash
	c.mu.Unlock()
	
	c.logger.Info("ENS name registered",
		zap.String("name", name))
	
	return nil
}

// UpdateENS обновляет ENS запись
func (c *ENSClient) UpdateENS(ctx context.Context, name string, newContentHash string) error {
	c.logger.Info("Updating ENS record",
		zap.String("name", name),
		zap.String("new_hash", newContentHash))
	
	// Проверка владения
	if _, ok := c.ensNames[name]; !ok {
		return fmt.Errorf("ENS name not found: %s", name)
	}
	
	// TODO: Вызов контракта Resolver для обновления
	
	c.mu.Lock()
	c.ensNames[name] = newContentHash
	c.mu.Unlock()
	
	c.logger.Info("ENS record updated",
		zap.String("name", name))
	
	return nil
}

// ResolveENS разрешает ENS имя в content hash
func (c *ENSClient) ResolveENS(ctx context.Context, name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	contentHash, ok := c.ensNames[name]
	if !ok {
		return "", fmt.Errorf("ENS name not found: %s", name)
	}
	
	return contentHash, nil
}

// GetENSInfo получает информацию о ENS имени
func (c *ENSClient) GetENSInfo(ctx context.Context, name string) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	contentHash, ok := c.ensNames[name]
	if !ok {
		return nil, fmt.Errorf("ENS name not found: %s", name)
	}
	
	// Получение информации из блокчейна
	// TODO: Вызов контрактов ENS
	
	return map[string]interface{}{
		"name":         name,
		"content_hash": contentHash,
		"owner":        c.publicAddress.Hex(),
		"registered":   true,
	}, nil
}

// CreateTransactor создаёт транзактор для подписи транзакций
func (c *ENSClient) CreateTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	
	transactor, err := bind.NewKeyedTransactorWithChainID(c.privateKey, chainID)
	if err != nil {
		return nil, err
	}
	
	// Настройка gas
	if c.config.GasPrice != nil {
		transactor.GasPrice = c.config.GasPrice
	}
	if c.config.GasLimit > 0 {
		transactor.GasLimit = c.config.GasLimit
	}
	
	return transactor, nil
}

// Start начинает фоновые задачи
func (c *ENSClient) Start(ctx context.Context) error {
	c.logger.Info("Starting ENS client")
	
	// Мониторинг событий ENS (опционально)
	go c.eventMonitor(ctx)
	
	return nil
}

// eventMonitor мониторит события ENS
func (c *ENSClient) eventMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Проверка изменений в ENS
			c.checkENSChanges()
		}
	}
}

// checkENSChanges проверяет изменения в ENS
func (c *ENSClient) checkENSChanges() {
	// TODO: Подписка на события NewOwner, NewResolver, etc.
}

// GetBalance получает баланс ETH
func (c *ENSClient) GetBalance(ctx context.Context) (*big.Int, error) {
	balance, err := c.client.BalanceAt(ctx, c.publicAddress, nil)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// WaitForTransaction ожидает подтверждения транзакции
func (c *ENSClient) WaitForTransaction(ctx context.Context, txHash common.Hash) error {
	// Используем прямой опрос по хэшу вместо bind.WaitMined для совместимости.
	receipt, err := c.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return err
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	c.logger.Info("Transaction confirmed",
		zap.String("hash", txHash.Hex()),
		zap.Uint64("block", receipt.BlockNumber.Uint64()))

	return nil
}
