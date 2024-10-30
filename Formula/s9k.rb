# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class S9k < Formula
  desc "A CLI tool for displaying AWS services"
  homepage "https://github.com/oslokommune/s9k"
  version "0.0.2"

  on_macos do
    on_intel do
      url "https://github.com/bsek/s9k/releases/download/v0.0.2/s9k_Darwin_x86_64.tar.gz"
      sha256 "23889473c81933eba93270db7f55be59a4faa00be236bdac74a244c8333bebe4"

      def install
        bin.install "s9k"
      end
    end
    on_arm do
      url "https://github.com/bsek/s9k/releases/download/v0.0.2/s9k_Darwin_arm64.tar.gz"
      sha256 "665699898b8cd3bfbebd155ed323966ac2c0f1fbae88d7f6313a81e757a94644"

      def install
        bin.install "s9k"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/bsek/s9k/releases/download/v0.0.2/s9k_Linux_x86_64.tar.gz"
        sha256 "a5e5c72dd950f9f78f46bdf9000d9912ca0c813100447ea19671660814c53756"

        def install
          bin.install "s9k"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/bsek/s9k/releases/download/v0.0.2/s9k_Linux_arm64.tar.gz"
        sha256 "530dcb55c707977e6102fa3c63d928c6983df1c4757608ea493804920750816c"

        def install
          bin.install "s9k"
        end
      end
    end
  end
end