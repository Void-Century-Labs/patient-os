import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function Home() {
  return (
    <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-8 px-6 py-16">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">PatientOS</h1>
        <p className="text-muted-foreground">
          Hospital queue and patient navigation platform — design system preview.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Join a Queue</CardTitle>
          <CardDescription>
            Sample form rendered with the shared design system components.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Full name</Label>
            <Input id="name" placeholder="Jane Doe" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="mobile">Mobile number</Label>
            <Input id="mobile" placeholder="+1 555 0100" />
          </div>
          <div className="flex items-center gap-2">
            <Badge>Cardiology</Badge>
            <Badge variant="secondary">Queue: 4</Badge>
            <Badge variant="success">Ready</Badge>
            <Badge variant="warning">Approaching</Badge>
          </div>
          <Button>Join Queue</Button>
        </CardContent>
      </Card>
    </main>
  );
}
