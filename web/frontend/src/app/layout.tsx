import { Button } from "@/components/ui/button";
import { storageController } from "@/storage";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Link, Outlet, useNavigate, useLocation } from "react-router";
import { ModeToggle } from "@/components/mode-toggle";
import { Github, Menu, Moon, Sun } from "lucide-react";
import Brandmark from "@/components/brandmark";
import { getPredictionModels, setPredictionModel } from "@/lib/set-prediction-model";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useMediaQuery } from "@/hooks/use-media-query";

export default function RootLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const isMobile = useMediaQuery("(max-width: 768px)"); // NOTE: Due to main page floating recommendation button, we need to add padding to the bottom of the page

  return (
    <>
      <div className="flex flex-col justify-center min-h-screen">
        <header className="bg-primary-foreground w-full">
          <div className="flex justify-between text-primary-background p-4 max-w-screen-lg mx-auto">
          <Link className="cursor-pointer flex items-center" to="/">
            <Brandmark className="inline-block h-8 max-w-none fill-foreground" />
          </Link>
          <div className="hidden sm:flex gap-2 items-center">
            <div className="flex rounded-md overflow-hidden divide-x">
              <Link
                to="/"
                className={`px-4 py-2 hover:bg-accent hover:text-accent-foreground ${location.pathname === "/" ? "underline" : ""}`}
              >
                Home
              </Link>
              <Link
                to="/visualization"
                className={`px-4 py-2 hover:bg-accent hover:text-accent-foreground ${location.pathname === "/visualization" ? "underline" : ""}`}
              >
                Visualization
              </Link>
              <Link
                to="/about"
                className={`px-4 py-2 hover:bg-accent hover:text-accent-foreground ${location.pathname === "/about" ? "underline" : ""}`}
              >
                About
            </Link>
            </div>
            <ResetButton
              onClick={() => {
                storageController.resetStorage();
                navigate("/");
                void (async () => {
                  await setPredictionModel(true);
                  window.location.reload();
                })();
              }}
            />
            <ModeToggle />
          </div>
          <div className="sm:hidden">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="p-1 rounded-md hover:bg-accent hover:text-accent-foreground">
                  <Menu className="h-6 w-6" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => navigate("/")}>
                  Home
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => navigate("/visualization")}>
                  Visualization
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => navigate("/about")}>
                  About
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <ResetButton
                    onClick={() => {
                      storageController.resetStorage();
                      navigate("/");
                      void (async () => {
                        await setPredictionModel(true);
                        window.location.reload();
                      })();
                    }}
                  >
                    <p className="w-full px-2 py-1 rounded-md hover:bg-accent hover:text-accent-foreground">Reset</p>
                  </ResetButton>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <div className="flex items-center justify-between">
                    <ModeToggle >
                      <div className="flex items-center gap-2">
                        <span>Theme:</span>
                        <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
                        <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
                      </div>
                    </ModeToggle>
                  </div>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          </div>
        </header>

        <div className="flex justify-center flex-1">
          <Outlet />
        </div>
      </div>
      <footer className={`bg-primary-foreground text-primary-background p-4 text-center ${isMobile && location.pathname === "/" ? "pb-20" : ""}`}>
        <a href="https://github.com/tomz197/pv254-project" target="_blank" rel="noreferrer">
          <Button variant="outline">GitHub <Github className="h-5 w-5" /></Button>
        </a>
        <div className="flex justify-center">
          <ModelSelector />
        </div>
      </footer>
    </>
  );
}

function ModelSelector() {
  const [selectedModel, setSelectedModel] = useState<string>(storageController.getPredictionModel() || "");
  const [open, setOpen] = useState(false);

  const modelsQuery = useQuery({
    queryKey: ["models"],
    queryFn: () => getPredictionModels(),
    staleTime: Infinity,
  });

  const handleModelChange = async (model: string) => {
    setSelectedModel(model);
    storageController.setPredictionModel(model);
    storageController.resetRelevance();
    setOpen(false);
    window.location.reload();
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger>
        <p className="text-muted-foreground text-xs mt-2">
          {selectedModel ? "Prediction model: " + selectedModel : "Select model"}
        </p>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <div className="flex flex-col gap-2">
          {modelsQuery.data?.map((model) => (
            <Button key={model} variant="ghost" onClick={() => handleModelChange(model)}>
              {model}
            </Button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}


function ResetButton({
  onClick,
  children,
}: {
  onClick: () => void,
  children?: React.ReactNode,
}) {
  return (
    <Dialog>
      <DialogTrigger asChild>
        {children ? children : <Button>Reset</Button>}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>This action will RESET algorithm</DialogTitle>
          <DialogDescription>
            By reseting algorithm, like and disliked courses will be removed with no way of returning them.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button onClick={onClick}>Reset</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
