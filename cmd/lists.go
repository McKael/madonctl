// Copyright © 2018 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/McKael/madon/v2"
)

var listsOpts struct {
	listID     int64
	accountID  int64
	accountIDs string
	title      string

	// Used for several subcommands to limit the number of results
	limit, keep uint
	all         bool
}

//listsCmd represents the lists command
var listsCmd = &cobra.Command{
	Use:     "lists",
	Aliases: []string{"list"},
	Short:   "Manage lists",
	Example: `  madonctl lists create --title "Friends"
  madonctl lists show
  madonctl lists show --list-id 3
  madonctl lists update --list-id 3 --title "Family"
  madonctl lists delete --list-id 3
  madonctl lists accounts --list-id 2
  madonctl lists add-accounts --list-id 2 --account-ids 123,456
  madonctl lists remove-accounts --list-id 2 --account-ids 456
  madonctl lists show --account-id 123`,
}

func init() {
	RootCmd.AddCommand(listsCmd)

	// Subcommands
	listsCmd.AddCommand(listsSubcommands...)

	listsCmd.PersistentFlags().UintVarP(&listsOpts.limit, "limit", "l", 0, "Limit number of API results")
	listsCmd.PersistentFlags().UintVarP(&listsOpts.keep, "keep", "k", 0, "Limit number of results")
	listsCmd.PersistentFlags().BoolVar(&listsOpts.all, "all", false, "Fetch all results")

	listsCmd.PersistentFlags().Int64VarP(&listsOpts.listID, "list-id", "G", 0, "List ID")

	listsGetSubcommand.Flags().Int64VarP(&listsOpts.accountID, "account-id", "a", 0, "Account ID number")
	// XXX accountUID?

	listsGetAccountsSubcommand.Flags().Int64VarP(&listsOpts.listID, "list-id", "G", 0, "List ID")

	listsCreateSubcommand.Flags().StringVar(&listsOpts.title, "title", "", "List title")
	listsUpdateSubcommand.Flags().StringVar(&listsOpts.title, "title", "", "List title")

	listsAddAccountsSubcommand.Flags().StringVar(&listsOpts.accountIDs, "account-ids", "", "Comma-separated list of account IDs")
	listsAddAccountsSubcommand.Flags().Int64VarP(&listsOpts.accountID, "account-id", "a", 0, "Account ID number")
	listsRemoveAccountsSubcommand.Flags().StringVar(&listsOpts.accountIDs, "account-ids", "", "Comma-separated list of account IDs")
	listsRemoveAccountsSubcommand.Flags().Int64VarP(&listsOpts.accountID, "account-id", "a", 0, "Account ID number")
}

var listsSubcommands = []*cobra.Command{
	listsGetSubcommand,
	listsCreateSubcommand,
	listsUpdateSubcommand,
	listsDeleteSubcommand,
	listsGetAccountsSubcommand,
	listsAddAccountsSubcommand,
	listsRemoveAccountsSubcommand,
}

var listsGetSubcommand = &cobra.Command{
	Use:   "show",
	Short: "Display one or several lists",
	// TODO Long: ``,
	Aliases: []string{"get", "display", "ls"},
	RunE:    listsGetRunE,
}

var listsGetAccountsSubcommand = &cobra.Command{
	Use:   "accounts --list-id N",
	Short: "Display a list's accounts",
	RunE:  listsGetAccountsRunE,
}

var listsCreateSubcommand = &cobra.Command{
	Use:   "create --title TITLE",
	Short: "Create a list",
	RunE:  listsSetDeleteRunE,
}

var listsUpdateSubcommand = &cobra.Command{
	Use:   "update --list-id N --title TITLE",
	Short: "Update a list",
	RunE:  listsSetDeleteRunE,
}

var listsDeleteSubcommand = &cobra.Command{
	Use:     "delete --list-id N",
	Short:   "Delete a list",
	Aliases: []string{"rm", "del"},
	RunE:    listsSetDeleteRunE,
}

var listsAddAccountsSubcommand = &cobra.Command{
	Use:     "add-accounts --list-id N --account-ids ACC1,ACC2...",
	Short:   "Add one or several accounts to a list",
	Aliases: []string{"add-account"},
	RunE:    listsAddRemoveAccountsRunE,
}

var listsRemoveAccountsSubcommand = &cobra.Command{
	Use:     "remove-accounts --list-id N --account-ids ACC1,ACC2...",
	Short:   "Remove one or several accounts from a list",
	Aliases: []string{"remove-account"},
	RunE:    listsAddRemoveAccountsRunE,
}

func listsGetRunE(cmd *cobra.Command, args []string) error {
	opt := listsOpts

	// Log in
	if err := madonInit(true); err != nil {
		return err
	}

	// Set up LimitParams
	var limOpts *madon.LimitParams
	if opt.all || opt.limit > 0 {
		limOpts = new(madon.LimitParams)
		limOpts.All = opt.all
	}
	if opt.limit > 0 {
		limOpts.Limit = int(opt.limit)
	}

	var obj interface{}
	var err error

	if opt.listID > 0 {
		var list *madon.List
		list, err = gClient.GetList(opt.listID)
		obj = list
	} else {
		var lists []madon.List
		lists, err = gClient.GetLists(opt.accountID, limOpts)

		if opt.keep > 0 && len(lists) > int(opt.keep) {
			lists = lists[:opt.keep]
		}
		obj = lists
	}

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	if obj == nil {
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
	}
	return p.printObj(obj)
}

func listsGetAccountsRunE(cmd *cobra.Command, args []string) error {
	opt := listsOpts

	if opt.listID <= 0 {
		return errors.New("missing list ID")
	}

	// Log in
	if err := madonInit(true); err != nil {
		return err
	}

	// Set up LimitParams
	var limOpts *madon.LimitParams
	if opt.all || opt.limit > 0 {
		limOpts = new(madon.LimitParams)
		limOpts.All = opt.all
	}
	if opt.limit > 0 {
		limOpts.Limit = int(opt.limit)
	}

	var obj interface{}
	var err error

	var accounts []madon.Account
	accounts, err = gClient.GetListAccounts(opt.listID, limOpts)

	if opt.keep > 0 && len(accounts) > int(opt.keep) {
		accounts = accounts[:opt.keep]
	}
	obj = accounts

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	if obj == nil {
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
	}
	return p.printObj(obj)
}

func listsSetDeleteRunE(cmd *cobra.Command, args []string) error {
	const (
		actionUnknown = iota
		actionCreate
		actionUpdate
		actionDelete
	)

	var action int
	opt := listsOpts

	switch cmd.Name() {
	case "create":
		if opt.listID > 0 {
			return errors.New("list ID should not be provided with create")
		}
		action = actionCreate
	case "update":
		if opt.listID <= 0 {
			return errors.New("list ID is required")
		}
		action = actionUpdate
	case "delete", "rm", "del":
		action = actionDelete
	}

	// Additionnal checks
	if action == actionUnknown {
		// Shouldn't happen.  If it does, might be an unrecognized alias.
		return errors.New("listsSetDeleteRunE: internal error")
	}

	if action != actionDelete && opt.title == "" {
		return errors.New("the list title is required")
	}

	// Log in
	if err := madonInit(true); err != nil {
		return err
	}

	var obj interface{}
	var err error
	var list *madon.List

	switch action {
	case actionCreate:
		list, err = gClient.CreateList(opt.title)
		obj = list
	case actionUpdate:
		list, err = gClient.UpdateList(opt.listID, opt.title)
		obj = list
	case actionDelete:
		err = gClient.DeleteList(opt.listID)
		obj = nil
	}

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	if obj == nil {
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
	}
	return p.printObj(obj)
}

func listsAddRemoveAccountsRunE(cmd *cobra.Command, args []string) error {
	opt := listsOpts

	if opt.listID <= 0 {
		return errors.New("missing list ID")
	}

	var ids []int64
	var err error
	ids, err = splitIDs(opt.accountIDs)
	if err != nil {
		return errors.New("cannot parse account IDs")
	}

	if opt.accountID > 0 { // Allow --account-id
		ids = []int64{opt.accountID}
	}
	if len(ids) < 1 {
		return errors.New("missing account IDs")
	}

	// Log in
	if err := madonInit(true); err != nil {
		return err
	}

	switch cmd.Name() {
	case "add-account", "add-accounts":
		err = gClient.AddListAccounts(opt.listID, ids)
	case "remove-account", "remove-accounts":
		err = gClient.RemoveListAccounts(opt.listID, ids)
	default:
		// Shouldn't happen.  If it does, might be an unrecognized alias.
		return errors.New("listsAddRemoveAccountsRunE: internal error")
	}

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}

	return nil
}
