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
    @Parameter(names = {"--alda-fingerprint"},
               description = "Used to identify this as an Alda process",
               hidden = true)
    public boolean aldaFingerprint = false;

    @Parameter(names = {"-h", "--help"},
               help = true,
               description = "Print this help text")
    public boolean help = false;

    @Parameter(names = {"-v", "--verbose"},
               description = "Enable verbose output")
    public boolean verbose = false;

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
  }

  private static class AldaCommand {
      @Parameter(names = {"--help", "-h"},
                 help = true,
                 hidden = true,
                 description = "Print this help text")
      public boolean help = false;

  }

  @Parameters(commandDescription = "Start the Alda server in the foreground.",
              hidden = true)
  private static class CommandServer extends AldaCommand {}

  @Parameters(commandDescription = "Start an interactive Alda REPL session.")
  private static class CommandRepl extends AldaCommand {}

  @Parameters(commandDescription = "Display this help text")
  private static class CommandHelp extends AldaCommand {}

  @Parameters(commandDescription = "Attempt to autoupdate alda")
  private static class CommandUpdate extends AldaCommand {}

  @Parameters(commandDescription = "Start the Alda server")
  private static class CommandStart extends AldaCommand {}

  @Parameters(commandDescription = "Stop the Alda server")
  private static class CommandStop extends AldaCommand {
    @Parameter (names = {"-y", "--yes"},
                description = "Auto-respond 'y' to confirm discarding " +
                              "unsaved changes")
    public boolean autoConfirm = false;
  }

  @Parameters(commandDescription = "Restart the Alda server")
  private static class CommandRestart extends AldaCommand {
    @Parameter (names = {"-y", "--yes"},
                description = "Auto-respond 'y' to confirm discarding " +
                              "unsaved changes")
    public boolean autoConfirm = false;
  }

  @Parameters(commandDescription = "List running Alda servers.")
  private static class CommandList extends AldaCommand {}

  @Parameters(commandDescription = "Display whether the server is up")
  private static class CommandStatus extends AldaCommand {}

  @Parameters(commandDescription = "Display the version of the Alda server")
  private static class CommandVersion extends AldaCommand {}

  @Parameters(commandDescription = "Display information about the Alda server")
  private static class CommandInfo extends AldaCommand {}

  @Parameters(commandDescription = "Evaluate and play Alda code")
  private static class CommandPlay extends AldaCommand {
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

    @Parameter(names = {"-a", "--append"},
               description = "Append to the existing score")
    public boolean appendToScore = false;

    @Parameter(names = {"-y", "--yes"},
               description = "Auto-respond 'y' to confirm e.g. score replacement")
    public boolean autoConfirm = false;
  }

  @Parameters(commandDescription = "Display the result of parsing Alda code")
  private static class CommandParse extends AldaCommand {
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

    @Parameter(names = {"-y", "--yes"},
               description = "Auto-respond 'y' to confirm e.g. score replacement")
    public boolean autoConfirm = false;
  }

  @Parameters(commandDescription = "Evaluate Alda code and append it to the " +
                                   "score without playing it")
  private static class CommandAppend extends AldaCommand {
    @Parameter(names = {"-f", "--file"},
               description = "Read Alda code from a file",
               converter = FileConverter.class)
    public File file;

    @Parameter(names = {"-c", "--code"},
               description = "Supply Alda code as a string")
    public String code;
  }

  @Parameters(commandDescription = "Display the score in progress")
  private static class CommandScore extends AldaCommand {
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
  private static class CommandNew extends AldaCommand {
    @Parameter(names = {"-f", "--file"},
               description = "A filename for the new score",
               converter = FileConverter.class)
    public File file;

    @Parameter(names = {"-y", "--yes"},
               description = "Auto-respond 'y' to confirm discarding " +
                             "unsaved changes")
    public boolean autoConfirm = false;
  }

  @Parameters(commandDescription = "Load a score from a file or string")
  private static class CommandLoad extends AldaCommand {
    @Parameter(names = {"-f", "--file"},
               description = "A file containing an Alda score",
               converter = FileConverter.class)
    public File file;

    @Parameter(names = {"-c", "--code"},
               description = "A string of Alda code")
    public String code;

    @Parameter(names = {"-y", "--yes"},
               description = "Auto-respond 'y' to confirm discarding " +
                             "unsaved changes")
    public boolean autoConfirm = false;
  }

  @Parameters(commandDescription = "Save the score to a file")
  private static class CommandSave extends AldaCommand {
    @Parameter(names = {"-f", "--file"},
               description = "The path to a file to which to save the score",
               converter = FileConverter.class)
    public File file;

    @Parameter(names = {"-y", "--yes"},
               description = "Auto-respond 'y' to confirm overwriting an " +
                             "existing file")
    public boolean autoConfirm = false;
  }

  @Parameters(commandDescription = "Edit the score")
  private static class CommandEdit extends AldaCommand {
    @Parameter(names = {"-e", "--editor"},
               description = "pass the file to a custom command instead of " +
                             "$EDITOR")
    public String editor;
  }

  public static void handleCommandSpecificHelp(JCommander jc, String name, AldaCommand c) {
      if(c.help) {
          jc.usage(name);
          System.exit(0);
      }
  }

  public static void main(String[] argv) {
    GlobalOptions globalOpts = new GlobalOptions();

    CommandHelp    help      = new CommandHelp();
    CommandUpdate  update    = new CommandUpdate();
    CommandServer  serverCmd = new CommandServer();
    CommandRepl    repl      = new CommandRepl();
    CommandStart   start     = new CommandStart();
    CommandStop    stop      = new CommandStop();
    CommandRestart restart   = new CommandRestart();
    CommandList    list      = new CommandList();
    CommandStatus  status    = new CommandStatus();
    CommandVersion version   = new CommandVersion();
    CommandInfo    info      = new CommandInfo();
    CommandPlay    play      = new CommandPlay();
    CommandParse   parse     = new CommandParse();
    CommandAppend  append    = new CommandAppend();
    CommandScore   score     = new CommandScore();
    CommandNew     newScore  = new CommandNew();
    CommandLoad    load      = new CommandLoad();
    CommandSave    save      = new CommandSave();
    CommandEdit    edit      = new CommandEdit();

    JCommander jc = new JCommander(globalOpts);
    jc.setProgramName("alda");

    jc.addCommand("help", help);

    jc.addCommand("update", update);

    jc.addCommand("server", serverCmd);
    jc.addCommand("repl", repl);

    jc.addCommand("start", start, "up", "init");
    jc.addCommand("stop", stop, "down");
    jc.addCommand("restart", restart, "downup");

    jc.addCommand("list", list);
    jc.addCommand("status", status);
    jc.addCommand("version", version);
    jc.addCommand("info", info);

    jc.addCommand("play", play);
    jc.addCommand("parse", parse);
    jc.addCommand("append", append, "add");

    jc.addCommand("score", score);
    jc.addCommand("new", newScore, "delete");
    jc.addCommand("load", load, "open");
    jc.addCommand("save", save);
    jc.addCommand("edit", edit);

    try {
      jc.parse(argv);
    } catch (ParameterException e) {
      System.out.println(e.getMessage());
      System.out.println();
      System.out.println("For usage instructions, see --help.");
      System.exit(1);
    }

    AldaServer server = new AldaServer(globalOpts.host,
                                       globalOpts.port,
                                       globalOpts.preBuffer,
                                       globalOpts.postBuffer);

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
          System.exit(0);

        case "update":
          handleCommandSpecificHelp(jc, "update", update);
          Util.updateAlda();
          System.exit(0);

        case "server":
          handleCommandSpecificHelp(jc, "server", serverCmd);
          server.startFg();
          break;

        case "repl":
          handleCommandSpecificHelp(jc, "repl", repl);
          server.startRepl();
          break;

        case "start":
        case "up":
        case "init":
          handleCommandSpecificHelp(jc, "start", start);
          server.startBg();
          System.exit(0);

        case "stop":
        case "down":
          handleCommandSpecificHelp(jc, "stop", stop);
          server.stop(stop.autoConfirm);
          System.exit(0);

        case "restart":
        case "downup":
          handleCommandSpecificHelp(jc, "restart", restart);
          server.restart(restart.autoConfirm);
          System.exit(0);

        case "list":
          handleCommandSpecificHelp(jc, "list", list);
          Util.listServers();
          System.exit(0);

        case "status":
          handleCommandSpecificHelp(jc, "status", status);
          server.status();
          System.exit(0);

        case "version":
          handleCommandSpecificHelp(jc, "version", version);
          System.out.println("Client version: " + Util.version());
          System.out.println();
          System.out.println("Server version:");
          server.version();
          System.exit(0);

        case "info":
          handleCommandSpecificHelp(jc, "info", info);
          server.info();
          System.exit(0);

        case "play":
          handleCommandSpecificHelp(jc, "play", play);
          inputType = Util.inputType(play.file, play.code);

          switch (inputType) {
            case "score":
              server.play();
              break;
            case "file":
              server.play(play.file, play.appendToScore, play.autoConfirm);
              break;
            case "code":
              server.play(play.code, play.appendToScore, play.autoConfirm);
              break;
            case "stdin":
              server.play(Util.getStdIn(), play.appendToScore, play.autoConfirm);
              break;
          }
          System.exit(0);

        case "parse":
          handleCommandSpecificHelp(jc, "parse", parse);
          mode = Util.scoreMode(parse.showLispCode, parse.showScoreMap);
          inputType = Util.inputType(parse.file, parse.code);

          switch (inputType) {
            case "file":
              server.parse(parse.file, mode, parse.autoConfirm);
              break;
            case "code":
              server.parse(parse.code, mode, parse.autoConfirm);
              break;
            case "stdin":
              server.parse(Util.getStdIn(), mode, parse.autoConfirm);
              break;
            default:
              throw new Exception("Please provide some Alda code in the form " +
                                  "of a string, file, or STDIN.");
          }
          System.exit(0);

        case "append":
        case "add":
          handleCommandSpecificHelp(jc, "append", append);
          inputType = Util.inputType(append.file, append.code);

          switch (inputType) {
            case "file":
              server.append(append.file);
              break;
            case "code":
              server.append(append.code);
              break;
            case "stdin":
              server.append(Util.getStdIn());
              break;
            default:
              throw new Exception("Please provide some Alda code in the form " +
                                  "of a string, file, or STDIN.");
          }
          System.exit(0);

        case "score":
          if(score.help) {
              jc.usage("score");
              System.exit(0);
          }
          mode = Util.scoreMode(score.showScoreText,
                                score.showLispCode,
                                score.showScoreMap);
          server.score(mode);
          System.exit(0);

        case "new":
        case "delete":
          handleCommandSpecificHelp(jc, "new", newScore);
          server.delete(newScore.autoConfirm);
          if (newScore.file != null) {
            server.save(newScore.file, newScore.autoConfirm);
          }
          System.exit(0);

        case "load":
        case "open":
          handleCommandSpecificHelp(jc, "load", load);
          inputType = Util.inputType(load.file, load.code);

          switch (inputType) {
            case "file":
              server.load(load.file, load.autoConfirm);
              break;
            case "code":
              server.load(load.code, load.autoConfirm);
              break;
            case "stdin":
              server.load(Util.getStdIn(), load.autoConfirm);
              break;
            default:
              throw new Exception("Please provide some Alda code in the form " +
                                  "of a string, file, or STDIN.");
          }
          System.exit(0);

        case "save":
          handleCommandSpecificHelp(jc, "save", save);
          inputType = Util.inputType(save.file, null);

          switch (inputType) {
            // "score" is returned when no filename is supplied
            case "score":
              server.save();
              break;
            case "file":
              server.save(save.file, save.autoConfirm);
              break;
            case "stdin":
              File file = new File(Util.getStdIn().trim());
              server.save(file, save.autoConfirm);
              break;
          }
          System.exit(0);

        case "edit":
          handleCommandSpecificHelp(jc, "edit", edit);
          String editor = System.getenv("EDITOR");
          if (edit.editor != null) {
            editor = edit.editor;
          }
          if (editor == null) {
            throw new Exception("EDITOR environment variable is not set.");
          }

          AldaServerInfo scoreInfo = server.getInfo();

          if (scoreInfo.filename == null) {
            server.msg("Score has not been saved yet. There is no file to " +
                       "edit.");
            break;
          }

          File file = new File(scoreInfo.filename);

          if (scoreInfo.isModified) {
            boolean saveFirst = Util.promptForConfirmation(
              "Your score has unsaved changes that will be lost unless you " +
              "save first.\nWould you like to save before editing?", false);

            if (saveFirst) {
              server.save();
            }
          }

          Util.runProgramInFg(editor, scoreInfo.filename);
          server.loadWithoutAsking(file);
          System.exit(0);
      }
    } catch (Exception e) {
      server.error(e.getMessage());
      if (globalOpts.verbose) {
        System.out.println();
        e.printStackTrace();
      }
      System.exit(1);
    }
  }

}
