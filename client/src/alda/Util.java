package alda;

public final class Util {

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

}
