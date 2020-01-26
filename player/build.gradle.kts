plugins {
  id("org.jetbrains.kotlin.jvm") version("1.3.60")
  application
}

repositories {
  mavenLocal()
  mavenCentral()
}

dependencies {
  implementation(kotlin("stdlib-jdk8"))
  implementation("com.illposed.osc:javaosc-core")
  testImplementation(kotlin("test"))
  testImplementation(kotlin("test-junit"))
}

dependencies {
  constraints {
    implementation("com.illposed.osc:javaosc-core:0.7-SNAPSHOT")
  }
}

application {
  mainClassName = "io.alda.player.MainKt"
}

val run by tasks.getting(JavaExec::class) {
  standardInput = System.`in`
}

val fatJar = task("fatJar", type = Jar::class) {
  archiveBaseName.set("${project.name}-fat")
  duplicatesStrategy = DuplicatesStrategy.INCLUDE
  manifest {
    attributes["Main-Class"] = "io.alda.player.MainKt"
  }
  from(configurations.compileClasspath.get().map({ if (it.isDirectory) it else zipTree(it) }))
  from(configurations.runtimeClasspath.get().map({ if (it.isDirectory) it else zipTree(it) }))
  with(tasks["jar"] as CopySpec)
}

tasks {
  "build" {
    dependsOn(fatJar)
  }
}
