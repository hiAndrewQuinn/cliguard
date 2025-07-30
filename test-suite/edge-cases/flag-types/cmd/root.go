package cmd

import (
	"net"
	"time"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command demonstrating all flag types
func NewRootCmd() *cobra.Command {
	// Basic types
	var stringFlag string
	var intFlag int
	var int8Flag int8
	var int16Flag int16
	var int32Flag int32
	var int64Flag int64
	var uintFlag uint
	var uint8Flag uint8
	var uint16Flag uint16
	var uint32Flag uint32
	var uint64Flag uint64
	var float32Flag float32
	var float64Flag float64
	var boolFlag bool

	// Special types
	var durationFlag time.Duration
	var ipFlag net.IP
	var ipMaskFlag net.IPMask
	var ipNetFlag net.IPNet
	var bytesHexFlag []byte
	var bytesBase64Flag []byte
	var countFlag int

	// Slice types
	var stringSliceFlag []string
	var intSliceFlag []int
	var int32SliceFlag []int32
	var int64SliceFlag []int64
	var uintSliceFlag []uint
	var float32SliceFlag []float32
	var float64SliceFlag []float64
	var boolSliceFlag []bool
	var durationSliceFlag []time.Duration
	var ipSliceFlag []net.IP

	// Map types
	var stringToStringFlag map[string]string
	var stringToInt64Flag map[string]int64

	rootCmd := &cobra.Command{
		Use:   "flagtypes",
		Short: "Test CLI for all flag types",
		Long:  "This CLI demonstrates all possible Cobra flag types for comprehensive testing.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("All flag types test CLI")
		},
	}

	// Basic type flags
	rootCmd.Flags().StringVarP(&stringFlag, "string", "s", "default", "String flag")
	rootCmd.Flags().IntVarP(&intFlag, "int", "i", 42, "Int flag")
	rootCmd.Flags().Int8Var(&int8Flag, "int8", 8, "Int8 flag")
	rootCmd.Flags().Int16Var(&int16Flag, "int16", 16, "Int16 flag")
	rootCmd.Flags().Int32Var(&int32Flag, "int32", 32, "Int32 flag")
	rootCmd.Flags().Int64Var(&int64Flag, "int64", 64, "Int64 flag")
	rootCmd.Flags().UintVarP(&uintFlag, "uint", "u", 42, "Uint flag")
	rootCmd.Flags().Uint8Var(&uint8Flag, "uint8", 8, "Uint8 flag")
	rootCmd.Flags().Uint16Var(&uint16Flag, "uint16", 16, "Uint16 flag")
	rootCmd.Flags().Uint32Var(&uint32Flag, "uint32", 32, "Uint32 flag")
	rootCmd.Flags().Uint64Var(&uint64Flag, "uint64", 64, "Uint64 flag")
	rootCmd.Flags().Float32Var(&float32Flag, "float32", 3.14, "Float32 flag")
	rootCmd.Flags().Float64Var(&float64Flag, "float64", 3.14159, "Float64 flag")
	rootCmd.Flags().BoolVarP(&boolFlag, "bool", "b", false, "Bool flag")

	// Special type flags
	rootCmd.Flags().DurationVarP(&durationFlag, "duration", "d", 5*time.Second, "Duration flag")
	rootCmd.Flags().IPVar(&ipFlag, "ip", net.IPv4(127, 0, 0, 1), "IP address flag")
	rootCmd.Flags().IPMaskVar(&ipMaskFlag, "ipmask", net.IPv4Mask(255, 255, 255, 0), "IP mask flag")
	rootCmd.Flags().IPNetVar(&ipNetFlag, "ipnet", net.IPNet{}, "IP network flag")
	rootCmd.Flags().BytesHexVar(&bytesHexFlag, "bytes-hex", []byte{0xDE, 0xAD, 0xBE, 0xEF}, "Bytes hex flag")
	rootCmd.Flags().BytesBase64Var(&bytesBase64Flag, "bytes-base64", []byte("hello"), "Bytes base64 flag")
	rootCmd.Flags().CountVarP(&countFlag, "count", "c", "Count flag (can be repeated)")

	// Slice type flags
	rootCmd.Flags().StringSliceVar(&stringSliceFlag, "string-slice", []string{"a", "b", "c"}, "String slice flag")
	rootCmd.Flags().IntSliceVar(&intSliceFlag, "int-slice", []int{1, 2, 3}, "Int slice flag")
	rootCmd.Flags().Int32SliceVar(&int32SliceFlag, "int32-slice", []int32{1, 2, 3}, "Int32 slice flag")
	rootCmd.Flags().Int64SliceVar(&int64SliceFlag, "int64-slice", []int64{1, 2, 3}, "Int64 slice flag")
	rootCmd.Flags().UintSliceVar(&uintSliceFlag, "uint-slice", []uint{1, 2, 3}, "Uint slice flag")
	rootCmd.Flags().Float32SliceVar(&float32SliceFlag, "float32-slice", []float32{1.1, 2.2, 3.3}, "Float32 slice flag")
	rootCmd.Flags().Float64SliceVar(&float64SliceFlag, "float64-slice", []float64{1.1, 2.2, 3.3}, "Float64 slice flag")
	rootCmd.Flags().BoolSliceVar(&boolSliceFlag, "bool-slice", []bool{true, false, true}, "Bool slice flag")
	rootCmd.Flags().DurationSliceVar(&durationSliceFlag, "duration-slice", []time.Duration{time.Second, time.Minute}, "Duration slice flag")
	rootCmd.Flags().IPSliceVar(&ipSliceFlag, "ip-slice", []net.IP{net.IPv4(127, 0, 0, 1)}, "IP slice flag")

	// Map type flags
	rootCmd.Flags().StringToStringVar(&stringToStringFlag, "string-to-string", map[string]string{"key": "value"}, "String to string map flag")
	rootCmd.Flags().StringToInt64Var(&stringToInt64Flag, "string-to-int64", map[string]int64{"key": 42}, "String to int64 map flag")

	// Persistent flags of different types
	rootCmd.PersistentFlags().String("persistent-string", "persistent", "Persistent string flag")
	rootCmd.PersistentFlags().Int("persistent-int", 100, "Persistent int flag")
	rootCmd.PersistentFlags().Bool("persistent-bool", false, "Persistent bool flag")

	// Required flags
	rootCmd.Flags().String("required-flag", "", "This flag is required")
	rootCmd.MarkFlagRequired("required-flag")

	// Deprecated flags
	rootCmd.Flags().String("deprecated-flag", "", "This flag is deprecated")
	rootCmd.Flags().MarkDeprecated("deprecated-flag", "use --new-flag instead")

	// Hidden flags
	rootCmd.Flags().String("hidden-flag", "", "This flag is hidden")
	rootCmd.Flags().MarkHidden("hidden-flag")

	// Add a subcommand to test flag inheritance
	rootCmd.AddCommand(newTestSubCmd())

	return rootCmd
}

func newTestSubCmd() *cobra.Command {
	var localFlag string
	var localInt int

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test subcommand",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Test subcommand executed")
		},
	}

	cmd.Flags().StringVar(&localFlag, "local-string", "local", "Local string flag")
	cmd.Flags().IntVar(&localInt, "local-int", 42, "Local int flag")

	return cmd
}