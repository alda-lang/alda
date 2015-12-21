package alda;

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

  public static void validateOpts(String file, String code) throws InvalidOptionsException {
    if (file == null && code == null) {
      throw new InvalidOptionsException("You must supply either a --file or " +
                                        "--code argument.");
    }

    if (file != null && code != null) {
      throw new InvalidOptionsException("You must supply either a --file or " +
                                        "--code argument (not both).");
    }
  }

  public static String getProgramPath() throws java.net.URISyntaxException {
    return Client.class.getProtectionDomain().getCodeSource().getLocation()
                 .toURI().getPath();
  }

  public static void forkProgram(Object... args) throws java.net.URISyntaxException, java.io.IOException {
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
