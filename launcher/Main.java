package launcher;

import java.lang.Runtime;
import java.io.File;
import java.io.IOException;
import java.lang.ProcessBuilder;
import java.lang.ProcessBuilder.Redirect;

public class Main {
    public static void main(final String[] args) throws IOException, InterruptedException {
        String executable = "minecraft-server-app";
        if (args.length > 0) {
            executable = args[0];
        }

        File cd = new File(System.getProperty("user.dir"));
        try {
            Runtime.getRuntime().exec("chmod +x "+cd.getAbsolutePath()+"/mc");
            Runtime.getRuntime().exec("chmod +x "+cd.getAbsolutePath()+"/rcon-cli");
            Runtime.getRuntime().exec("chmod +x "+cd.getAbsolutePath()+"/"+executable);
        } catch (Exception ex) {
            ex.printStackTrace();
        }

        ProcessBuilder pb = new ProcessBuilder(cd.getAbsolutePath()+"/"+executable);
        pb.directory(cd);
        pb.redirectOutput(Redirect.INHERIT);
        pb.redirectError(Redirect.INHERIT);

        Process ps = pb.start();
        System.exit(ps.waitFor());
    }
}
