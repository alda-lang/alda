package alda;

import java.io.BufferedReader;
import java.io.Console;
import java.io.File;
import java.io.InputStreamReader;
import java.io.InputStream;
import java.io.IOException;
import java.net.URISyntaxException;
import java.net.URL;
import java.net.HttpURLConnection;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.Scanner;

import clojure.java.api.Clojure;
import clojure.lang.IFn;
import clojure.lang.ISeq;
import clojure.lang.Symbol;
import clojure.lang.ArraySeq;
import com.google.gson.*;

import java.net.MalformedURLException;
import java.io.BufferedInputStream;
import java.io.BufferedInputStream;
import java.io.FileOutputStream;

import org.apache.commons.io.FileUtils;
import org.apache.commons.lang3.SystemUtils;

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

  public static boolean promptForConfirmation(String prompt) {
    Console console = System.console();
    if (System.console() != null) {
      Boolean confirm = null;
      while (confirm == null) {
        String response = console.readLine(prompt + " (y/n) ");
        if (response.toLowerCase().startsWith("y")) {
          confirm = true;
        } else if (response.toLowerCase().startsWith("n")) {
          confirm = false;
        }
      }
      return confirm.booleanValue();
    } else {
      System.out.println(prompt + "\n");
      System.out.println("Unable to get a response because you are " +
                         "redirecting input.\nI'm just gonna assume the " +
                         "answer is no.\n\n" +
                         "To auto-respond yes, use the -y/--yes option.");
      return false;
    }
  }

  public static String inputType(File file, String code)
    throws InvalidOptionsException {
    if (file == null && code == null) {
      // check to see if we're receiving input from STDIN
      if (System.console() == null) {
        return "stdin";
      } else {
        // if not, input type is the existing score in its entirety
        return "score";
      }
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

  public static String getStdIn() {
    String fromStdIn = "";
    Scanner scanner = new Scanner(System.in);
    while (scanner.hasNextLine()) {
      fromStdIn += scanner.nextLine();
    }
    return fromStdIn;
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
    return Main.class.getProtectionDomain().getCodeSource().getLocation()
               .toURI().getPath();
  }

  public static String makeApiCall(String apiRequest) throws IOException {
      URL url = new URL(apiRequest);
      HttpURLConnection conn =
        (HttpURLConnection) url.openConnection();

      if (conn.getResponseCode() != 200) {
        throw new IOException(conn.getResponseMessage());
      }

      // Buffer the result into a string
      BufferedReader rd = new BufferedReader(
        new InputStreamReader(conn.getInputStream()));
      StringBuilder sb = new StringBuilder();
      String line;
      line = rd.readLine();
      while (line != null) {
        sb.append(line);
        line = rd.readLine();
      }
      rd.close();
      conn.disconnect();
      return sb.toString();
  }

  public static void downloadFile(String url, String path) {
    BufferedInputStream in = null;
    FileOutputStream fout = null;
    try {
      in = new BufferedInputStream(new URL(url).openStream());
      fout = new FileOutputStream(path);

      final byte data[] = new byte[1024];
      int count;
      while ((count = in.read(data, 0, 1024)) != -1) {
        fout.write(data, 0, count);
      }
    } catch (MalformedURLException e) {
      System.err.println("An error occured while downloading a file (1).");
      e.printStackTrace();
    } catch (IOException e) {
      System.err.println("An error occured while downloading a file (2).");
      e.printStackTrace();
    }finally {
      // Close file IO's
      try {
        if (in != null) {
          in.close();
        }
        if (fout != null) {
          fout.close();
        }
      } catch (IOException e) {
        // We can't do anything.
        System.err.println("A critical error occured while downoading a file (3).");
        e.printStackTrace();
        return;
      }
    }
  }

  public static String readFile(File file) throws IOException {
    return FileUtils.readFileToString(file, StandardCharsets.UTF_8);
  }

  public static String readResourceFile(String path) {
    StringBuilder out = new StringBuilder();
    BufferedReader reader = null;
    try {
      InputStream in = Util.class.getClassLoader().getResourceAsStream(path);
      reader = new BufferedReader(new InputStreamReader(in));
      String line;
      while ((line = reader.readLine()) != null) {
        out.append(line);
      }
    } catch(IOException e) {
      System.err.println("There was an error reading a file!");
      e.printStackTrace();
    } finally {
      try {
        reader.close();
      } catch (Exception e) {
        // Theres nothing we can do...
        System.err.println("There was a critical error!");
        e.printStackTrace();
        return null;
      }
    }
    return out.toString();
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

  public static void runProgramInFg(String... args)
  throws IOException, InterruptedException {
    new ProcessBuilder(args).inheritIO().start().waitFor();
  }

  public static void callClojureFn(String fn, Object... args) {
    Symbol var = (Symbol)Clojure.read(fn);
    IFn require = Clojure.var("clojure.core", "require");
    require.invoke(Symbol.create(var.getNamespace()));
    ISeq argsSeq = ArraySeq.create(args);
    Clojure.var(var.getNamespace(), var.getName()).applyTo(argsSeq);
  }
}
