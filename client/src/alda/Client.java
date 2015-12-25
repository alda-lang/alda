package alda;

import java.io.File;

import com.beust.jcommander.IStringConverter;
import com.beust.jcommander.JCommander;
import com.beust.jcommander.Parameter;
import com.beust.jcommander.Parameters;
import com.beust.jcommander.ParameterException;

public class Client {

  public static class FileConverter implements IStringConverter<File> {
    @Override
    public File convert(String value) {
      return new File(value);
    }
  }

  private static class GlobalOptions {
    @Parameter(names = {"-h", "--help"},
               help = true,
               description = "Print this help text")
    public boolean help = false;

    @Parameter(names = {"-H", "--host"},
               description = "The hostname of the Alda server")
    public String host = "localhost";

    @Parameter(names = {"-p", "--port"},
               description = "The port of the Alda server")
    public int port = 27713;

    @Parameter(names = {"-b", "--pre-buffer"},
               description = "A number of milliseconds of lead time for buffering")
    public int preBuffer = 0;

    @Parameter(names = {"-B", "--post-buffer"},
               description = "A number of milliseconds to wait after playing " +
                             "the score, before exiting")
    public int postBuffer = 1000;

    @Parameter(names = {"-s", "--stock"},
               description = "Use the default MIDI soundfont of your JVM, " +
                             "instead of FluidR3")
    public boolean useStockSoundfont = false;
  }

  @Parameters(commandDescription = "Start the Alda server in the foreground.",
              hidden = true)
  private static class CommandServer {}

  @Parameters(commandDescription = "Start an interactive Alda REPL session.")
  private static class CommandRepl {}

  @Parameters(commandDescription = "Display this help text")
  private static class CommandHelp {}

  @Parameters(commandDescription = "Start the Alda server")
  private static class CommandStart {}

  @Parameters(commandDescription = "Stop the Alda server")
  private static class CommandStop {}

  @Parameters(commandDescription = "Restart the Alda server")
  private static class CommandRestart {}

  @Parameters(commandDescription = "Display whether the server is up")
  private static class CommandStatus {}

  @Parameters(commandDescription = "Display the version of the Alda server")
  private static class CommandVersion {}

  @Parameters(commandDescription = "Evaluate and play Alda code")
  private static class CommandPlay {
    @Parameter(names = {"-f", "--file"},
               description = "Read Alda code from a file",
               converter = FileConverter.class)
    public File file;

    @Parameter(names = {"-c", "--code"},
               description = "Supply Alda code as a string")
    public String code;

    @Parameter(names = {"-F", "--from"},
               description = "A time marking or marker from which to start " +
                             "playback")
    public String from;

    @Parameter(names = {"-T", "--to"},
               description = "A time marking or marker at which to end playback")
    public String to;

    @Parameter(names = {"-r", "--replace"},
               description = "Replace the existing score with new code")
    public boolean replaceScore = false;
  }

  @Parameters(commandDescription = "Display the score in progress")
  private static class CommandScore {
    @Parameter(names = {"-t", "--text"},
               description = "Display the score text")
    public boolean showScoreText = false;

    @Parameter(names = {"-l", "--lisp"},
               description = "Display the score in the form of alda.lisp " +
                             "(Clojure) code")
    public boolean showLispCode = false;

    @Parameter(names = {"-m", "--map"},
               description = "Display the map of score data")
    public boolean showScoreMap = false;
  }

  @Parameters(commandDescription = "Delete the score and start a new one")
  private static class CommandNew {}

  @Parameters(commandDescription = "Edit the score in progress")
  private static class CommandEdit {
    @Parameter(names = {"-e", "--editor"},
               description = "pass the file to a custom command instead of " +
                             "$EDITOR")
    public String editor;
  }

  @Parameters(commandDescription = "Display the result of parsing Alda code")
  private static class CommandParse {
    @Parameter(names = {"-f", "--file"},
               description = "Read Alda code from a file",
               converter = FileConverter.class)
    public File file;

    @Parameter(names = {"-c", "--code"},
               description = "Supply Alda code as a string")
    public String code;

    @Parameter(names = {"-l", "--lisp"},
               description = "Display the score in the form of alda.lisp " +
                             "(Clojure) code")
    public boolean showLispCode = false;

    @Parameter(names = {"-m", "--map"},
               description = "Display the map of score data")
    public boolean showScoreMap = false;
  }

  public static void main(String[] argv) {
    GlobalOptions globalOpts = new GlobalOptions();

    CommandHelp help        = new CommandHelp();
    CommandServer serverCmd = new CommandServer();
    CommandRepl repl        = new CommandRepl();
    CommandStart start      = new CommandStart();
    CommandStop stop        = new CommandStop();
    CommandRestart restart  = new CommandRestart();
    CommandStatus status    = new CommandStatus();
    CommandVersion version  = new CommandVersion();
    CommandPlay play        = new CommandPlay();
    CommandScore score      = new CommandScore();
    CommandNew newScore     = new CommandNew();
    CommandEdit edit        = new CommandEdit();
    CommandParse parse      = new CommandParse();

    JCommander jc = new JCommander(globalOpts);
    jc.setProgramName("alda");

    jc.addCommand("server", serverCmd);
    jc.addCommand("repl", repl);

    jc.addCommand("start", start, "up", "init");
    jc.addCommand("stop", stop, "down");
    jc.addCommand("restart", restart, "downup");

    jc.addCommand("status", status);
    jc.addCommand("version", version);
    jc.addCommand("help", help);

    jc.addCommand("play", play);
    jc.addCommand("score", score);
    jc.addCommand("new", newScore, "delete");
    jc.addCommand("edit", edit);
    jc.addCommand("parse", parse);

    try {
      jc.parse(argv);
    } catch (ParameterException e) {
      System.out.println(e.getMessage());
      System.out.println();
      System.out.println("For usage instructions, see --help.");
      System.exit(1);
    }

    AldaServer server = new AldaServer(globalOpts.host, globalOpts.port,
                                       globalOpts.preBuffer, globalOpts.postBuffer,
                                       globalOpts.useStockSoundfont);

    try {
      if (globalOpts.help) {
        jc.usage();
        return;
      }

      String command = jc.getParsedCommand();
      command = command == null ? "help" : command;

      // used for play and parse commands
      String mode;
      String inputType;

      switch (command) {
        case "help":
          jc.usage();
          break;

        case "server":
          server.startFg();
          break;

        case "repl":
          server.startRepl();
          break;

        case "start":
        case "up":
        case "init":
          server.startBg();
          break;
        case "stop":
        case "down":
          server.stop();
          break;
        case "restart":
        case "downup":
          server.restart();
          break;

        case "status":
          server.status();
          break;
        case "version":
          server.version();
          break;

        case "play":
          inputType = Util.inputType(play.file, play.code);

          switch (inputType) {
            case "score":
              server.play();
              break;
            case "file":
              server.play(play.file, play.replaceScore);
              break;
            case "code":
              server.play(play.code, play.replaceScore);
              break;
          }
          break;
        case "score":
          mode = Util.scoreMode(score.showScoreText,
                                score.showLispCode,
                                score.showScoreMap);
          server.score(mode);
          break;
        case "new":
        case "delete":
          server.delete();
          break;
        case "edit":
          // TODO
          break;
        case "parse":
          mode = Util.scoreMode(parse.showLispCode, parse.showScoreMap);
          inputType = Util.inputType(parse.file, parse.code);

          switch (inputType) {
            case "file":
              server.parse(parse.file, mode);
              break;
            case "code":
              server.parse(parse.code, mode);
              break;
          }
          break;
      }
    } catch (Exception e) {
      server.error(e.getMessage());
      System.exit(1);
    }
  }

}

