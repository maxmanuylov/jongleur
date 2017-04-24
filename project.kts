import org.intellij.cluster.project.dsl.*

object version {
    val major = 0
    val minor = 9
}

project {
    val versionText = "v${version.major}.${version.minor}"
    val buildText = "$versionText.0.$buildNumber"

    make {
        on("maxmanuylov/go-build:1.8") {
            at("/go/src/github.com/maxmanuylov/jongleur")
            withEnv {
                + "VERSION".to(versionText)
                + "BUILD".to(buildText)
                + "REVISION".to(vcsRevision)
            }
            run("/bin/bash", "build/build.sh")
        }
    }
}
