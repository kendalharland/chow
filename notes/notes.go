package notes

import (
	"encoding/json"
	"testing"
)

// chow aims to provide these major advantages over the existing recipes framework:
// 1. Static references to steps, so that mocking input is much easier.
// 2. Static, Strong Types.
// 3. Filesystem checks and gaurantees.
// 4. Direct references to filesystem entities (instead of just string paths).
// 5. Ability to use files without Placeholders.

// A step needs to provide 3 things to the framework:
// 1. A specification of how to create its shell command
// 2. The set of arguments it expects as input
// 3. The set of outputs it generates.
//
// The user should not be able to specify extra computation within a step. A step should
// basically just be a template that values are plugged into.   For example, a user should
// not be able to define a step-within-a-step.
//
// It's important that the framework understand the names of the step's inputs, and their
// types.
//
// Why are these things important? Because if we allow the user to do complex computation
// inside a step defintion, then steps are no longer this atomic, single shell command
// template that we want them to be.
//
// How can we make the user specify the input arguments and their types, and a
// template for creating the actual shell command in the step, without allowing the user
// to do complex logic within a step? Well, that depends on how we declare what a step is.

type I_Step interface{}

// If we make an interface called "Step", the idea is that we some part of the framework
// consumes this interface and does what it needs to do. Since we want steps to exist as a
// sort of data-container, This doesn't really fit without our model of what a step is.

type S_Step struct {
	// The set of inputs to this Step template. These are the things callers will pass when
	// invoking your template.
	Inputs interface{} // Or struct maybe?

	// The set of command line arguments generated in this step.
	Command []string
	Outputs interface{}
}

// If we make a struct called "Step" this is close to the data-container model that we're
// aiming for.  To specify input properties, we can ask the user to embed another struct
// within the step, whose values are fetched via reflection. As an example:

var echoUsernameStep = S_Step{
	// Username: The name of the user.
	Inputs: struct {
		Username string
	}{},
	Outputs: struct {
	}{},

	// Run "echo" with the name of the user.
	Command: []string{"echo", "{.Username}"},
}

// This also lets us define the step result in a similar fashion.

// One problem with defining the step this way is that the client of your step libary
// can re-assign your step identifiers. It stinks that Go doesn't disallow this.

type F_Step func(username string)

// If we design Step as a function, it's not clear how the user supplies the outputs and
// the command line arguments. It's also not clear how we use reflection to retrieve the
// set of inputs. Futhermore this approach allows the user to specify extra code in the
// body of the step, which is stricly against what we want.

// If there is no other option, it seems like we should design the Step as struct. But
// what's the best way to use structs for this purpose? We could declare a single
// data-class like we did above. We could also make the syntax a bit more neat by simply
// requiring the user to define their own struct and pass it to some register function:

type FileReference int
type DirReference int

// XXX_Inputs and XXX_Outputs are separate from the XXX step itself so that callers can
// create instances of them when mocking, consume filepaths, and calling the step.

type HelloStepInputs struct {
	Username string
}
type HelloStepOutputs struct {
	MyFile FileReference
}
type HelloStep_T struct {
	Inputs  HelloStepInputs
	Outputs HelloStepOutputs
	Command string `exec:"echo 'Hello, {.Username}!"`
}

var HelloStep = HelloStep_T{}

// If you don't want anyone to specify arguments to your step, or you don't want to assert
// any outputs (or maybe your step just don't produce any) then it's ok to inline these
// things.
type NoContractStep struct {
	Inputs  struct{ Foo string }
	Outputs struct{}
	Command string `exec:"no one consumes this output"`
}

/* If we instead provide the user with a contract on the type of struct they're allowed to
* pass in to the framework, we get all of the benefits of Go's static type system without
* doing any crazy hacking.  This also prevents clients of your library from redefining or
* reassigning the steps in your package. For now, this certainly seems better than the
* above approach.  We'll stick with this style until we hit a roadblack that we can't
* overcome.
*
* So this declaration style gives us static typing, which is great. Where do we go from
* here? Well, once the user has defined their steps, the framework needs to verify that
* the step has the proper fields. Since we can't validate this statically, we can do it
* at compile time, and even provide a harness for testing that all step definitions are
* valid.
*
* Clients will need to invoke your step somehow. To do that, they should refer to your
* step by-name.  Since your step is not a function, we'll need to provide some way to do
* that.
 */

func Run(step interface{}) error {
	return nil
}
func Run2(step interface{}, inputs interface{}) error {
	return nil
}

func TestRun() {
	_ = Run(HelloStep_T{Inputs: HelloStepInputs{
		Username: "kendal",
	}})

	_ = Run2(HelloStep, HelloStepInputs{
		Username: "kendal",
	})
}

/* We need to represent a step result. A step result would have the same type for every
* step, were it not for a step's outputs, which we need to be able to mock for tests.
* Maybe we can ignore this for now.
 */

type StepResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Outputs  interface{}
}

/* You can always safely type-case StepResult.Outputs to the type XXX_Outputs.  This is a
* Guarantee the framework gives you.
*
* The framework also gaurantees that every output in StepResult.Outputs exists on the
* local filesystem.  If not, the process will have crashed before the step result was
* returned.
* Let's take a look at how mocking step results should work.
*
* A mock specifies how the framework should create StepResults when a specific step
* invocation occurs. The invocation is identified by
 */
type When struct {
	Called     interface{}
	Count      int
	WithInputs interface{}
}
type Return struct {
	Stdout  string
	Stderr  string
	Outputs interface{}
}
type Mock struct {
	When   When
	Return Return
}

/** Interface implementation -------------------------------------------------------------
*
* I ultimately decided to go with the interface + implementation of a step rather than
* having users define step "templates" as pure structs for the following reasons:
* 1. Types in Go are not first-class objects, so users can't create a step template's
*    "Inputs" struct when invoking a step.  This also means that user's can't reference
*    A step for mocking without creating an instance of the step and passing that to the
*    framework.  this is awkard because there is extra information in the struct that
*    isn't used, but you get all of this when you instantiate it.
* 2. There's no easy way to encode commad arguments when your step template is just a
*    struct.
* 3. Command arguments cannot always be declared, sometimes they must be computed e.g. if
*    certain things about the input arguments or filesystem are different.
*
* Things to keep in mind when taking this approach:
* 1. Do not have a global chow instance. That way, StepProviders can't run themselves
*    Inside their `Provide` method, which would all people to do really complex things
*    inside steps.
* 2. function types can implement interfaces, which makes mocking via reflection difficult,
*    because we must figure out how to mock functions and struct types. As an example,
*    Foo below implements the Step interface:
 */
type Command int
type Outputs int

type Step interface {
	Provide() (Command, Outputs)
}

type Foo func() (Command, Outputs)

func (f *Foo) Provide() (Command, Outputs) {
	return (*f)()
}

/**
 * For this reason, we actually won't use map lookups to compare step invocations, because
 * it's complicated and error prone to write equality checks and hashes. A simple linear
 * search works instead.
 */

/** How to do tests
 *
 * Step invocations in tests are not actually executed as shell commands.  Instead the
 * shell commands are recorded with the given inputs in "expection files".  Recipes does
 * not have the concept of outputs, so we need to design how those work. But first, let's
 * focus on step invocations.
 *
 * What do we need to do with step invocations in tests?
 * 1. Log them in expectation files
 * 2. Mock their outputs/results
 * 3. Return zero values for them if no mocks are present.
 *
 * Assuming we stick with the current struct-based method for mocking step results, we
 * still need to figure out how to do the following:
 * 1. Log step invocations
 * 2. Associate mocks with step invocations
 *
 * How should we log step invocations? Recipes chooses some predefined directory and
 * defines a JSON file there with the recipe's name (or recipe module's) combined with the
 * test case name.  We can follow this pattern, but how do we log to that file?
 *
 * To keep most of the code flexible, the engine internals should just dump formatted data
 * to an io.Writer.  Some other "coordinating" part of the engine can handle creating a
 * writer that goes to the appropriate file.
 */

/** Caveats of implement mocks as a struct type.
 *
 * Up to this point I've been sticking with this implementation of a mock object:
 *
 *   type When Step
 *   type Return StepResult
 *   type Mock struct {
 *     When   When
 *     Return Return
 *     Count int
 *   }
 *
 * The problem with this structure is that once the user passes a Mock object to the
 * framework, we have no way of knowing whether the zero values in the Mock object were
 * left empty or explicitly set by the user.  This approach falls short because of a
 * feature of the Go language.  Instead we need to give the user the ability to express
 * that certain fields should be left as-is, while others should be overriden.
 *
 * For example, the Mock declaration `Mock{&Foo{a=1}}` where Foo is a
 * struct { a, b, c, int } could mean many things:
 *
 * 1. Mock any call to Foo where a == 1 and b and c are anything
 * 2. Mock any call to Foo where a == 1 and b and c are the zero value (0)
 * 3. Mock any call to Foo where a == 1 and b is anything and c is the zero value (0)
 *
 * Most often, the user probably wants option 1.  But we still have to disambiguate
 * between 1, 2, and 3.  We can adopt a simple rule to continue solving this problem: By
 * default, assume all zero-value inputs can be anything.  After this, we can figure out
 * how to enable to the user to explicity mark certain inputs as "the zero value" rather
 * than "anything" to solve this problem.
 *
 * Option one: Give the user some constant that can be assigned to the value they're
 * indicating must be a zero value.
 *
 * Option one(a): Give the user a value wrapped as the type they require. The internal
 * value can be retrieved later on by the framework. For example:
 */

type zeroInt int

var zint int = int(zeroInt(0))

/**
 * ^^^ This is impossible to do because we can't recover type information after the user
 *     casts the value to another type (unless it's an interface)

 * Option two: ???
 *
 * It might be the case that users usually don't want separte conditions for each of a, b,
 * and c, and that most of the time they only need to express one of the following:
 *
 * 1. Mock any call to Foo where a == 1 and b and c are anything
 * 2. Mock any call to Foo where a == 1 and b and c are the zero value (0)
 *
 * This binary problem is easier to solve: we can just pass a boolean to the mock object
 * specifying whether to treat all zero values as "anything" or "zero".  This boolean
 * should default to false so that case 1 is the default.
 */

/** What are outputs for?
 *
 * I had this idea that the filesystem entities generated by a step should be declared in
 * some way, so that the framework can provide certain feedback to the programmer about
 * the files and directories they attempt to read/write:
 * 1. Warn when reading from a file that doesn't exist
 * 2. Make certain assertions testable - such as Python's `os.path.exists()`.
 * 3. Error when a command does not produce a file that it should have, rather than
 *    continuing to execute.
 * 4. Warn when writing to a file that already exists.
 *
 * It's complicated to think about how to support all of these use-cases.  I think (1), (2)
 * and (3) are the most useful, so we can focus on those.
 *
 * Both (1) and (3) are things that can be solved by assertions. e.g.:
 *
 *
 * func make_file(path string ) {  make the file.  }
 * func run() {
 *     make_file( "file.txt")
 *     assert path.Exists("file.txt")
 * }
 *
 */

/**
 * So is it really all that useful to bake the concept of outputs into the framework,
 * rather than just making assertions testable?  I guess we kind of have to, because
 * we can't assert that a file exists during testing if each step doesn't provide a
 * record of the files that should exist.
 *
 * Ok then, so steps should definitely have the ability to provide this list of files,
 * and assertions should be checked by the framework in tests and fail-hard in prod
 * if they fail.  That gets us (2) and (3).
 *
 * The environment we're running in determines how we treat outputs and path
 * assertions:
 *
 * In production:
 * 1. We error if a command did not produce a file it claimed to produce.
 * 2. We error if a user's path assertion fails
 *
 * In tests:
 * 1. We error if a user's path assertion cannot be satisfied because no step declares
 *    the path.
 *
 * This gets us:
 * 1. Testable fielpath assertions
 * 2. Safe execution in prod, since missing outputs are check right after they should
 *    have been created.
 * 3. The ability to assert relative paths exist.
 *
 * When a User's assertion that some path exists fails during testing, we should print
 * a helpful error message in both prod and in testing.  What should the user do when
 * they get an error message saying that we can't garuantee that a file will exist,
 * because no step declares it?
 *
 * 1. Send a patch upstream to declare the output? what are the consequences of doing this?
 */

/** How do we declare outputs?
 *
 */

/** Questions about paths in recipes.
 * 1. Why do we use special path roots like [START_DIR]?
 * 2. What are the problems that I have with paths in recipes?
 */

/** Why paths are a problem when declaring outputs.
 *
 * Expectation files should not contain any information about absolute paths, because
 * paths are not portable.  Expectation files generated in one directory wil not be
 * the same as expectation files generated in another directory. Recipes handles this
 * providing a set of global path names that you should use in production and tests.
 * Can we copy that approach? How could it be improved?
 *
 *  1. It's hard to discover paths because they're strings, so once implemented, they
 *     must still be documented somewhere in order for users to know about them.
 *  2. Recipes provides a library for working with paths, but it's confusing to use
 *     because it only exports a subset of the core Python path library, and it's not
 *     implemented to work well with tests (certain functions are no-ops in tests).
 *  3. Overall the *strategy* that recipes-py used to work with paths is probably ok
 *     for most people's needs.  It was just executed poorly.
 *
 * The paths that users want most often are probably the current task root, and the
 * current working directory.  Let's start with these and let the implementation evolve
 * naturally from there.
 *
 * One caveat of copying recipes' approach is that we *have* to provide libraries for
 * changing directories and spoofing paths if we want everything to be testable. The user
 * can't just go around calling os.Chdir! And when they do, we should at least *pretend*
 * like we changed directories.
 *
 * Example: How should this code in testing and production?
 *
 *   # Assume we start at root
 *   mkdir('./foo')
 *   cd('./foo')
 *   mkdir('./bar')
 *   assertExists('//foo/bar')
 *
 * If we didn't keep track of the cwd when testing, we'd think that this code created 2
 * top-level directories named 'foo' and 'bar', so `assertExists('//foo/bar`) would fail.
 *
 * The root of the problem is that changing directories is a system operation that alters
 * the global state of the task, just like writing files. We have to spoof it somehow. We
 * will probably also need to provide libraries for working with paths.  When designing
 * this library we should always keep the convenience of the user in mind.  It should be
 * extremely trivial to write code that modifies paths, the working dir, etc that works in
 * both production and tests.
 *
 * NOTE: The test step detail logger should probably convert all paths to absolute paths
 *       before logging, right?
 *
 * NOTE: Paths should always be provided in Unix syntax.  The framework should convert
 *       them to the appropriate version for the platform at runtime. There's nothing more
 *       annoying than having to account for this in your code.
 */

/* Difference between giving a the user a function to get the absolute path to the cwd or
   root vs giving the user a function to create an absolute path from a path string:

   abs = api.absoluteOf("./path/to/file")

   vs

   abs = api.cwd() + "path/to/file"

   1. In the first example, we need a complex path resolution function in `absoluteOf`.
      for testing.  In prod we can defer to an os specific call.
   2. If we don't allow the user to just use ./, then we must make the path API available
	  to every part of our program so that things like Steps can call api.cwd to create
	  their outputs. (This is the main benefit of strings: they're always globally
	  available)

	The second option seems to offer the most convenience to the user. We should stick
	with that.

	So we'll have complex code that understands certain strings. How will this work?
	In production, we should map each string to it's real world equivalent.  In testing
	for now we can leave them as-is just to get a better idea of what the output looks
	like, and if it needs to change.
*/

/** Steps need path conversion
 *
 * Steps will need an api for converting paths before they are passed to commands.  The
 * framework should probably handle this itself since Steps don't execute commands
 * directly.
 *
 * We need some way of detecting when a path has been passed to a command. We should also
 * give the user some way of specifying that all they want is a file, and they don't care
 * what path that file has. This is what placeholders are for!  We can't use placeholders
 * in Go though, because we don't want to support dynamic types in dart arguments; Only
 * strings.
 *
 * Can we do this entirely through strings? The danger with strings is that the user might
 * pass a string value that matches the string we propose as a placeholder, which would
 * result in confusing behavior.
 *
 * Proposal 1 for detecting paths:
 * 1. Every single path passed to the framework must begin with a predefined prefix:
 *    a. // - for the start directory
 *    b. //CWD// - for the cwd
 *    c. //PLACEHOLDER// - for a placeholder
 *
 * 2. The can wrap the entire argument in quotes if they wish to prevent the framework
 *    from escaping the path
 *
 * 3. Using //PLACEHOLDER// as a the key for a placeholder means that the framework has to
 *    support some sort of object to refer to that placeholder later on.  Instead we
 *    should just make it easy to create a throwaway directory or file.  This is a big way
 *    in which we differ from recipes.
 *
 * 4. These path references should always be parsed from longest to shortest to avoid
 *    errors.
 */
// Path constants.
//
// TODO: document these somewhere.
const (
	// The directory in which execution started.  In production this is filled in with the
	// actual path to the execution directory.
	StartDir string = "//"

	// The current working directory.  At the start of execution, this is always equal to
	// Root.  The directory referenced by Cwd changes as the application changes
	// directories.
	PathCwd string = "//CWD//"
)

/** Do we need Placeholders? What are they for? Can we replace them with something else?
 *
 */

// This may come in handy

func formatJson(t *testing.T, js string) string {
	var container interface{}
	if err := json.Unmarshal([]byte(js), &container); err != nil {
		t.Fatal(err)
	}
	formatted, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(formatted)
}
