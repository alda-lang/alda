package alda;

import com.google.gson.*;

import com.jcabi.manifests.Manifests;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStreamReader;
import java.io.IOException;
import java.net.URISyntaxException;
import java.util.ArrayList;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.Scanner;

import org.apache.commons.lang3.SystemUtils;

public class AldaClient {
  public static String version() {
    return Manifests.read("alda-version");
  }

  public static void updateAlda() throws URISyntaxException {
    // Get the path to the current alda executable
    String programPath = Util.getProgramPath();
    String latestApiStr = "https://api.github.com/repos/alda-lang/alda/releases/latest";
    String apiResult;
    String clientVersion = version();

    // Make a call to the Github API to get the latest version number/download URL
    try {
      apiResult = Util.makeApiCall(latestApiStr);
    } catch (IOException e) {
      System.err.println("There was an error connecting to the Github API.");
      e.printStackTrace();
      return;
    }

    // Turn api result into version numbers and links
    Gson gson = new Gson();
    JsonObject job = gson.fromJson(apiResult, JsonObject.class);

    // Gets the download URL. This may have ...alda or ...alda.exe
    String downloadURL = null;
    String dlRegex = SystemUtils.IS_OS_WINDOWS ? ".*alda\\.exe$" : ".*.alda$";
    String latestTag = job.getAsJsonObject().get("tag_name").toString().replaceAll("\"", "");

    // Check to see if we currently have the version determined by latestTag
    if (latestTag.indexOf(clientVersion) != -1 || clientVersion.indexOf(latestTag) != -1) {
      System.out.println("Your version of alda (" + clientVersion +") is up to date!");
      return;
    }

    for (JsonElement i : job.getAsJsonArray("assets")) {
      String candidate = i.getAsJsonObject().get("browser_download_url").toString().replaceAll("\"", "");
      if (candidate.matches(dlRegex)) {
        downloadURL = candidate;
        break;
      }
    }

    if (downloadURL == null) {
      System.err.println("Alda download link not found for your platform.");
      return;
    }

    // Request confirmation from user:
    System.out.print("Install alda '" + latestTag + "' over '" + clientVersion + "' ? [yN]: ");
    System.out.flush();
    String name = (new Scanner(System.in)).nextLine();
    if (!(name.equalsIgnoreCase("y") || name.equalsIgnoreCase("yes"))) {
      System.out.println("Quitting...");
      return;
    }

    System.out.println("Downloading " + downloadURL + "...");

    // Download file from downloadURL to programPath
    Util.downloadFile(downloadURL, programPath);
    // set as executable if on UNIX
    if (SystemUtils.IS_OS_UNIX) {
      new File(programPath).setExecutable(true);
    }
    System.out.println();
    System.out.println("Updated alda " + clientVersion + " => " + latestTag + ".");
    System.out.println("If you have any currently running servers, you may want to restart them so that they are running the latest version.");
  }

  public static ArrayList<AldaProcess> findProcesses() throws IOException {
    ArrayList<AldaProcess> processes = new ArrayList<AldaProcess>();

    Process p = Runtime.getRuntime().exec("ps -e");
    InputStreamReader isr = new InputStreamReader(p.getInputStream());
    BufferedReader input = new BufferedReader(isr);
    String line;
    while ((line = input.readLine()) != null) {
      if (line.contains("alda-fingerprint")) {
        AldaProcess process = new AldaProcess();

        Matcher a = Pattern.compile("^\\s*(\\d+).*").matcher(line);
        Matcher b = Pattern.compile(".*--port (\\d+).*").matcher(line);
        Matcher c = Pattern.compile(".* server.*").matcher(line);
        Matcher d = Pattern.compile(".* worker.*").matcher(line);
        if (a.find()) {
          process.pid = Integer.parseInt(a.group(1));
          if (b.find()) {
            process.port = Integer.parseInt(b.group(1));
          } else {
            process.port = -1;
          }

          if (c.find()) {
            process.type = "server";
          }

          if (d.find()) {
            process.type = "worker";
          }
        }

        processes.add(process);
      }
    }
    input.close();
    p.getInputStream().close();
    p.getOutputStream().close();
    p.getErrorStream().close();
    p.destroy();
    return processes;
  }

  public static void listProcesses(int timeout) throws IOException {
    if (!SystemUtils.IS_OS_UNIX) {
      System.out.println("Sorry -- listing running processes is not " +
                         "currently supported for Windows users.");
      return;
    }

    ArrayList<AldaProcess> processes = findProcesses();
    for (AldaProcess process : processes) {
      if (process.type == "server") {
        if (process.port == -1) {
          System.out.printf("[???] Mysterious server running on unknown " +
                            "port (pid: %d)\n", process.pid);
          System.out.flush();
        } else {
          AldaServer server = new AldaServer("localhost",
                                             process.port,
                                             timeout,
                                             false);
          server.status();
        }
      } else if (process.type == "worker") {
        if (process.port == -1) {
          System.out.printf("[???] Mysterious worker running on unknown " +
                            "port (pid: %d)\n", process.pid);
          System.out.flush();
        } else {
          System.out.printf("[%d] Worker (pid: %d)\n", process.port, process.pid);
          System.out.flush();
        }
      } else {
        if (process.port == -1) {
          System.out.printf("[???] Mysterious Alda process running on " +
                            "unknown port (pid: %d)\n", process.pid);
          System.out.flush();
        } else {
          System.out.printf("[%d] Mysterious Alda process (pid: %d)\n",
                            process.port, process.pid);
          System.out.flush();
        }
      }
    }
  }

  public static boolean checkForExistingServer(int port) throws IOException {
    ArrayList<AldaProcess> processes = findProcesses();
    for (AldaProcess process : processes) {
      if (process.port == port) {
        return true;
      }
    }

    return false;
  }
}
