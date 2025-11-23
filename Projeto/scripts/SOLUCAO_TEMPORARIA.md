# Solução Temporária para o Problema "invalid sender"

## Problema
O Geth está rejeitando transações com erro "invalid sender" mesmo com a conta desbloqueada.

## Solução: Usar a Conta do Signer para Deploy

Como a conta do signer (0x792A31E1989e59c226bcfaf3E151BD95Ab5e625F) tem 1 milhão de ETH e está configurada no genesis, você pode:

### Opção 1: Usar a Conta do Signer no Cliente Go

1. Execute o cliente Go:
   ```batch
   cd scripts
   test-client.bat
   ```

2. Quando pedir para escolher conta, escolha a primeira conta (a do signer)
   - Senha: `123456`

3. Faça o deploy do contrato (opção 6)

### Opção 2: Transferir ETH Manualmente (quando o problema for resolvido)

Depois que conseguirmos fazer transações funcionarem, você pode transferir ETH da conta do signer para outras contas usando o script `fund-account.bat`.

## Próximos Passos

Precisamos resolver o problema "invalid sender" no Geth. Possíveis causas:
- Clique não está criando blocos automaticamente
- Geth precisa de configuração adicional
- Problema com a forma como as transações estão sendo enviadas

