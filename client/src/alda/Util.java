package alda;

import java.io.File;
import java.io.IOException;
import java.net.URISyntaxException;
import java.util.Arrays;

import clojure.java.api.Clojure;
import clojure.lang.IFn;
import clojure.lang.ISeq;
import clojure.lang.Symbol;
import clojure.lang.ArraySeq;

public final class Util {

  public static Object[] concat(Object[] a, Object[] b) {
    int aLen = a.length;
    int bLen = b.length;
    Object[] c = new Object[aLen+bLen];
    System.arraycopy(a, 0, c, 0, aLen);
    System.arraycopy(b, 0, c, aLen, bLen);
    return c;
  }

  public static Object[] conj(Object[] a, Object b) {
    return concat(a, new Object[]{b});
  }

  public static String inputType(File file, String code)
    throws InvalidOptionsException {
    if (file == null && code == null) {
      return "score";
    }

    if (file != null && code != null) {
      throw new InvalidOptionsException("You must supply either a --file or " +
                                        "--code argument (not both).");
    }

    if (file != null) {
      return "file";
    } else {
      return "code";
    }
  }

  public static String scoreMode(boolean showLispCode,
                                 boolean showScoreMap)
    throws InvalidOptionsException {
    boolean[] modes = { showLispCode, showScoreMap };
    int count = 0; for (boolean mode : modes) { if (mode) { count++; } }
    if (count > 1) {
      throw new InvalidOptionsException("You must choose either --lisp or " +
                                        "--map mode (not both).");
    } else if (count == 1) {
      if (showLispCode)  { return "lisp"; }
      if (showScoreMap)  { return "map"; }
    }

    // default to lisp mode if no options provided
    return "lisp";
  }

  public static String scoreMode(boolean showScoreText,
                                 boolean showLispCode,
                                 boolean showScoreMap)
    throws InvalidOptionsException {
    boolean[] modes = { showScoreText, showLispCode, showScoreMap };
    int count = 0; for (boolean mode : modes) { if (mode) { count++; } }
    if (count > 1) {
      throw new InvalidOptionsException("You must choose only one mode out " +
                                        "of --text, --lisp or --map.");
    } else if (count == 1) {
      if (showScoreText) { return "text"; }
      if (showLispCode)  { return "lisp"; }
      if (showScoreMap)  { return "map"; }
    }

    // default to text mode if no options provided
    return "text";
  }

  public static String getProgramPath() throws URISyntaxException {
    return Client.class.getProtectionDomain().getCodeSource().getLocation()
                 .toURI().getPath();
  }

  public static void forkProgram(Object... args)
    throws URISyntaxException, IOException {
    String programPath = getProgramPath();

    Object[] program;
    if (programPath.endsWith(".jar")) {
      program = new Object[]{"java", "-jar", programPath};
    } else {
      program = new Object[]{programPath};
    }

    Object[] objectArray = concat(program, args);
    String[] execArgs = Arrays.copyOf(objectArray, objectArray.length, String[].class);

    Runtime.getRuntime().exec(execArgs);
  }

  public static void callClojureFn(String fn, Object... args) {
    Symbol var = (Symbol)Clojure.read(fn);
    IFn require = Clojure.var("clojure.core", "require");
    require.invoke(Symbol.create(var.getNamespace()));
    ISeq argsSeq = ArraySeq.create(args);
    Clojure.var(var.getNamespace(), var.getName()).applyTo(argsSeq);
  }

}
