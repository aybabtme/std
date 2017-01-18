package flag

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func ParseSet(appname string, declaration func(FlagSet)) error {
	return parseSet(appname, os.Args, os.Environ(), declaration)
}

func parseSet(appname string, args, environs []string, declaration func(FlagSet)) error {
	fs := newFlagSet(appname, args, environs)
	declaration(fs)
	if err := fs.parse(); err != nil {
		return fs.errorf(err)
	}
	if len(args) == 2 {
		arg := strings.TrimPrefix(args[1], "-")
		arg = strings.TrimPrefix(arg, "-")

		if arg == "help" || arg == "h" {
			fmt.Fprintln(os.Stderr, fs.usage())
			os.Exit(1)
		}
	}
	return nil
}

type FlagSet interface {
	String(v *string, usage string, parts ...string)
	Int(v *int, usage string, parts ...string)
	Float64(v *float64, usage string, parts ...string)
	Bool(v *bool, usage string, parts ...string)
	Duration(v *time.Duration, usage string, parts ...string)

	// Time(v *time.Time, usage string, parts ...string)
	// Bytes(v *unit.Byte, usage string, parts ...string)
}

type flagSet struct {
	appname      string
	apropos      string
	declarations []string

	args    []string
	environ []string
	argfs   *flag.FlagSet
	envfs   *flag.FlagSet
}

func newFlagSet(appname string, args, environ []string) *flagSet {
	return &flagSet{
		appname: appname,
		args:    args,
		environ: environ,
		argfs:   flag.NewFlagSet(appname, flag.ContinueOnError),
		envfs:   flag.NewFlagSet(appname, flag.ContinueOnError),
	}
}

func (fs *flagSet) parse() error {
	// pretend the ENV_VARS are flags
	envargs := make([]string, 0, len(fs.environ))
	for _, value := range fs.environ {
		if fs.envfs.Lookup(value[:strings.Index(value, "=")]) == nil {
			continue
		}
		envargs = append(envargs, "-"+value)
	}
	// env vars by default
	if err := fs.envfs.Parse(envargs); err != nil {
		return err
	}
	// flags override the env vars
	return fs.argfs.Parse(fs.args)
}

func (fs *flagSet) errorf(err error) error {
	sort.Strings(fs.declarations)
	return fmt.Errorf("%s: bad usage, %v\n%s\n%s",
		fs.appname, err,
		fs.apropos,
		strings.Join(fs.declarations, "\n"),
	)
}

func (fs *flagSet) usage() string {
	sort.Strings(fs.declarations)
	return fmt.Sprintf("%s: %s\n%s",
		fs.appname, fs.apropos,
		strings.Join(fs.declarations, "\n"),
	)
}

func (fs *flagSet) String(v *string, usage string, parts ...string) {
	var value string
	if v != nil {
		value = *v
	}
	fs.argfs.StringVar(v, flagName(fs.appname, parts), value, usage)
	fs.envfs.StringVar(v, envName(fs.appname, parts), value, usage)
	fs.recordDeclaration(parts, value, "string", usage)
}

func (fs *flagSet) Int(v *int, usage string, parts ...string) {
	var value int
	if v != nil {
		value = *v
	}
	fs.argfs.IntVar(v, flagName(fs.appname, parts), value, usage)
	fs.envfs.IntVar(v, envName(fs.appname, parts), value, usage)
	fs.recordDeclaration(parts, value, "integer", usage)
}

func (fs *flagSet) Float64(v *float64, usage string, parts ...string) {
	var value float64
	if v != nil {
		value = *v
	}
	fs.argfs.Float64Var(v, flagName(fs.appname, parts), value, usage)
	fs.envfs.Float64Var(v, envName(fs.appname, parts), value, usage)
	fs.recordDeclaration(parts, value, "float", usage)
}

func (fs *flagSet) Bool(v *bool, usage string, parts ...string) {
	var value bool
	if v != nil {
		value = *v
	}
	fs.argfs.BoolVar(v, flagName(fs.appname, parts), value, usage)
	fs.envfs.BoolVar(v, envName(fs.appname, parts), value, usage)
	fs.recordDeclaration(parts, value, "boolean", usage)
}

func (fs *flagSet) Duration(v *time.Duration, usage string, parts ...string) {
	var value time.Duration
	if v != nil {
		value = *v
	}
	fs.argfs.DurationVar(v, flagName(fs.appname, parts), value, usage)
	fs.envfs.DurationVar(v, envName(fs.appname, parts), value, usage)
	fs.recordDeclaration(parts, value, "duration", usage)
}

func (fs *flagSet) recordDeclaration(parts []string, value interface{}, typename, usage string) {
	fs.declarations = append(fs.declarations, fmt.Sprintf("-%s=%v (%s) a %s, %s",
		flagName(fs.appname, parts),
		value,
		envName(fs.appname, parts),
		typename,
		usage,
	))
}

func flagName(appname string, name []string) string {
	lowerNames := mapfn(name, strings.ToLower)
	return strings.Join(lowerNames, ".")
}

func envName(appname string, name []string) string {
	in := append([]string{appname}, name...)
	upperNames := mapfn(in, strings.ToUpper)
	return strings.Join(upperNames, "_")
}

func mapfn(in []string, fn func(string) string) []string {
	out := make([]string, 0, len(in))
	for _, n := range in {
		out = append(out, fn(n))
	}
	return out
}
