package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"

	"github.com/CyberMiles/travis/server/commands"
	"github.com/CyberMiles/travis/vm/cmd/utils"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage accounts",
	Long: `
Manage accounts, list all existing accounts, import a private key into a new
account, create a new account or update an existing account.

It supports interactive mode, when you are prompted for password as well as
non-interactive mode where passwords are supplied via a given password file.
Non-interactive mode is only meant for scripted use on test networks or known
safe environments.

Make sure you remember the password you gave when creating a new account (with
either new or import). Without it you are not able to unlock your account.

Note that exporting your key in unencrypted format is NOT supported.

Keys are stored under <DATADIR>/keystore.
It is safe to transfer the entire directory or the individual keys therein
between nodes by simply copying.

Make sure you backup your keys regularly.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func prepareAccountCommands() {
	accountCmd.AddCommand(
		accountListCmd,
		accountCreateCmd,
		accountUpdateCmd,
		accountImportCmd,
	)

	fsAccount := pflag.NewFlagSet("", pflag.ContinueOnError)
	fsAccount.StringP(commands.FlagPassword, "p", "", "Password file to use for non-interactive password input")
	fsAccount.BoolP(commands.FlagLightKDF, "", false, "Reduce key-derivation RAM & CPU usage at some expense of KDF strength")

	accountCreateCmd.Flags().AddFlagSet(fsAccount)
	accountUpdateCmd.Flags().AddFlagSet(fsAccount)
	accountImportCmd.Flags().AddFlagSet(fsAccount)
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print summary of existing accounts",
	Long:  "Print a short summary of all accounts",
	RunE:  accountList,
}

var accountCreateCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new account",
	Long: `
Creates a new account and prints the address.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the --password flag.

Note, this is meant to be used for testing only, it is a bad idea to save your
password to file or expose in any other way.
`,
	RunE: accountCreate,
}

var accountUpdateCmd = &cobra.Command{
	Use:   "update <address>",
	Short: "Update an existing account",
	Long: `
Update an existing account.

The account is saved in the newest version in encrypted format, you are prompted
for a passphrase to unlock the account and another to save the updated file.

This same command can therefore be used to migrate an account of a deprecated
format to the newest format or change the password for an account.

For non-interactive use the passphrase can be specified with the --password flag.

Since only one password can be given, only format update can be performed,
changing your password is only possible interactively.
`,
	RunE: accountUpdate,
}

var accountImportCmd = &cobra.Command{
	Use:   "import <keyfile>",
	Short: "Import a private key into a new account",
	Long: `
Imports an unencrypted private key from <keyfile> and creates a new account.
Prints the address.

The keyfile is assumed to contain an unencrypted private key in hexadecimal format.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the --password flag.

Note:
As you can directly copy your encrypted accounts to another travis instance,
this import mechanism is not needed when you transfer an account between
nodes.
`,
	RunE: accountImport,
}

func accountList(cmd *cobra.Command, args []string) error {
	ctx, err := commands.SetupAccountContext()
	if err != nil {
		return err
	}
	stack := utils.MakeFullNode(ctx)
	var index int
	for _, wallet := range stack.AccountManager().Wallets() {
		for _, account := range wallet.Accounts() {
			fmt.Printf("Account #%d: {%x} %s\n", index, account.Address, &account.URL)
			index++
		}
	}
	return nil
}

// accountCreate creates a new account into the keystore defined by the CLI flags.
func accountCreate(cmd *cobra.Command, args []string) error {
	ctx, err := commands.SetupAccountContext()
	if err != nil {
		return err
	}
	stack := utils.MakeFullNode(ctx)

	keystore := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	password := utils.GetPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.",
		true, 0, ethUtils.MakePasswordList(ctx))
	address, err := keystore.NewAccount(password)

	if err != nil {
		ethUtils.Fatalf("Failed to create account: %v", err)
	}
	fmt.Printf("Address: {%x}\n", address.Address)
	return nil
}

// accountUpdate transitions an account from a previous format to the current
// one, also providing the possibility to change the pass-phrase.
func accountUpdate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		ethUtils.Fatalf("No accounts specified to update")
	}

	ctx, err := commands.SetupAccountContext()
	if err != nil {
		return err
	}
	stack := utils.MakeFullNode(ctx)

	keystore := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	for _, addr := range args {
		account, oldPassword := utils.UnlockAccount(ctx, keystore, addr, 0, nil)
		newPassword := utils.GetPassPhrase("Please give a new password. Do not forget this password.",
			true, 0, nil)
		if err := keystore.Update(account, oldPassword, newPassword); err != nil {
			ethUtils.Fatalf("Could not update the account: %v", err)
		}
	}
	return nil
}

func accountImport(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		ethUtils.Fatalf("keyfile must be given as argument")
	}
	key, err := crypto.LoadECDSA(args[0])
	if err != nil {
		ethUtils.Fatalf("Failed to load the private key: %v", err)
	}

	ctx, err := commands.SetupAccountContext()
	if err != nil {
		return err
	}
	stack := utils.MakeFullNode(ctx)
	passphrase := utils.GetPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.",
		true, 0, ethUtils.MakePasswordList(ctx))

	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	acct, err := ks.ImportECDSA(key, passphrase)
	if err != nil {
		ethUtils.Fatalf("Could not create the account: %v", err)
	}
	fmt.Printf("Address: {%x}\n", acct.Address)
	return nil
}
