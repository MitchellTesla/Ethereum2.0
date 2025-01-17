// +build e2e

package e2e

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/consul-terraform-sync/api"
	"github.com/hashicorp/consul-terraform-sync/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_MetaCOmmandErrors(t *testing.T) {
	// test cases that cross subcommands that coded in the command meta object
	t.Parallel()

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "meta_errs")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	port, err := api.FreePort()
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	config := fakeHandlerConfig().appendPort(port).
		appendConsulBlock(srv).appendTerraformBlock(tempDir)
	config.write(t, configPath)

	cmd, err := runSyncDevMode(configPath)
	defer stopCommand(cmd)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	cases := []struct {
		name           string
		args           []string
		outputContains string
	}{
		{
			"missing required arguments",
			[]string{},
			"Error: this command requires one argument",
		},
		{
			"unsupported argument",
			[]string{"-unsupported-flag", fakeFailureTaskName},
			"Error: unsupported arguments in flags",
		},
		{
			"non-existing task",
			[]string{fmt.Sprintf("-port=%d", port), "non-existent-task"},
			"does not exist or has not been initialized yet",
		},
		{
			"out of order arguments",
			[]string{fakeFailureTaskName, fmt.Sprintf("-port %d", port)},
			"All flags are required to appear before positional arguments",
		},
	}

	for _, lifecycle := range []string{"disable", "enable"} {
		// run through all the test cases for each task lifcycle command
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%s/%s", lifecycle, tc.name), func(t *testing.T) {
				subcmd := []string{"task", lifecycle}
				subcmd = append(subcmd, tc.args...)

				output, err := runSubcommand(t, "", subcmd...)
				assert.Contains(t, output, tc.outputContains)
				assert.Error(t, err)
			})
		}
	}

	delete()
}

func TestE2E_EnableTaskCommand(t *testing.T) {
	t.Parallel()

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "enable_cmd")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	port, err := api.FreePort()
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	config := disabledTaskConfig().appendPort(port).
		appendConsulBlock(srv).appendTerraformBlock(tempDir)
	config.write(t, configPath)

	cmd, err := runSyncDevMode(configPath)
	defer stopCommand(cmd)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	cases := []struct {
		name           string
		args           []string
		input          string
		outputContains string
	}{
		{
			"happy path",
			[]string{fmt.Sprintf("-port=%d", port), disabledTaskName},
			"yes\n",
			"enable complete!",
		},
		{
			"user does not approve plan",
			[]string{fmt.Sprintf("-port=%d", port), disabledTaskName},
			"no\n",
			"Cancelled enabling task",
		},
		{
			"help flag",
			[]string{"-help"},
			"",
			"Usage: consul-terraform-sync task enable [options] <task name>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subcmd := []string{"task", "enable"}
			subcmd = append(subcmd, tc.args...)

			output, err := runSubcommand(t, tc.input, subcmd...)
			assert.NoError(t, err)
			assert.Contains(t, output, tc.outputContains)
		})
	}

	delete()
}

func TestE2E_DisableTaskCommand(t *testing.T) {
	t.Parallel()

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "disable_cmd")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	port, err := api.FreePort()
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	config := baseConfig().appendPort(port).appendConsulBlock(srv).appendDBTask()
	config.write(t, configPath)

	cmd, err := runSyncDevMode(configPath)
	defer stopCommand(cmd)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	cases := []struct {
		name           string
		args           []string
		outputContains string
	}{
		{
			"happy path",
			[]string{fmt.Sprintf("-port=%d", port), dbTaskName},
			"disable complete!",
		},
		{
			"help flag",
			[]string{"-help"},
			"consul-terraform-sync task disable [options] <task name>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subcmd := []string{"task", "disable"}
			subcmd = append(subcmd, tc.args...)

			output, err := runSubcommand(t, "", subcmd...)
			assert.Contains(t, output, tc.outputContains)
			assert.NoError(t, err)
		})
	}

	delete()
}

// runSubcommand runs a CTS subcommand and its arguments. If user input is
// required for subcommand, pass it through 'input' parameter. Function returns
// the stdout/err output and any error when executing the subcommand.
// Note: if error returned, output will still contain any stdout/err information.
func runSubcommand(t *testing.T, input string, subcmd ...string) (string, error) {
	cmd := exec.Command("consul-terraform-sync", subcmd...)

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)

	err = cmd.Start()
	require.NoError(t, err)

	_, err = stdin.Write([]byte(input))
	require.NoError(t, err)
	stdin.Close()

	err = cmd.Wait()
	return b.String(), err
}
